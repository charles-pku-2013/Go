package system_metric

import (
    "log"
    // "fmt"
    // "sync"
    "math"
    "sort"
    "time"
    "sync/atomic"
)

// enum
const (
    LATENCY int32 = 0
    CPU           = 1
    MIX           = 2
    QPS           = 3
    INT_MAX       = 2147483647
)

type _ValueMap map[string]*float64
type PATTERN_TO_MAP map[int32]_ValueMap

type RateLimiterManager struct {
    run_                          bool

    // 共用参数
    update_qps_time_ms_           int32
    max_qps_curr_                 float64       // 当前最大qps
    current_position_             int32         // 队列的当前位置
    current_qps_                  int32         // atomic 统计当前qps
    current_qps_handle_           int32         // atomic 统计当前未被限流qps
    is_rate_limit_                bool          // 当前是否发生限流
    curr_sample_number_           int32         // 当前采样窗口大小

    // mix模式
    low_load_latency_             int64         // 低负载时latency，即cpu<min_cpu_usage
    high_load_latency_            int64         // 高负载时latency，即cpu=82~88
    high_load_qps_number_         int32
    high_load_qps_count_          int32
    high_load_qps_position_       int32
    high_load_qps_                int32         // 高负载时 qps 作为机器最高负载能力

    request_latency_vec_          []int64       // 保存最近 max_sample_number_ 个请求的latency
    high_load_qps_vec_            []int64       // 高负载时 qps ，选出最大值作为 high_qps

    // std::shared_ptr<CPUInfo> cpu_info_ = nullptr;  // NOTE 由docker_cpu.go提供
    rate_limiter_config_          *RateLimiterConfig
    rate_limiter_                 *TokenBucketRateLimiter
    rate_limiter_cpu_pattern_     *RateLimiterCPUPattern      // cpu模式
    rate_limiter_mix_pattern_     *RateLimiterMixPattern      // cpu和延迟时间混合模式
    rate_limiter_latency_pattern_ *RateLimiterLatencyPattern  // 延迟模式
    rate_limiter_qps_pattern_     *RateLimiterQPSPattern      // QPS模式

    rate_limiter_map_             PATTERN_TO_MAP              // 存放限流在线更改的参数
    rate_limiter_pattern_map_     map[string]int32
}

func NewRateLimiterManager() *RateLimiterManager {
    obj := new(RateLimiterManager)

    obj.run_ = true
    //  共用参数
    obj.update_qps_time_ms_ = 1000
    obj.max_qps_curr_ = 500.0
    obj.current_position_ = -1
    atomic.StoreInt32(&obj.current_qps_, 200)  // TODO 为什么初始值是200 ??
    atomic.StoreInt32(&obj.current_qps_handle_, 200)
    obj.is_rate_limit_ = false
    //  mix模式
    obj.low_load_latency_  = 0
    obj.high_load_latency_ = 0
    obj.high_load_qps_number_ = 4
    obj.high_load_qps_count_ = 0
    obj.high_load_qps_position_ = 0
    obj.high_load_qps_ = 1000

    obj.rate_limiter_pattern_map_ = map[string]int32{}
    obj.rate_limiter_pattern_map_["cpu"] = CPU
    obj.rate_limiter_pattern_map_["latency"] = LATENCY
    obj.rate_limiter_pattern_map_["mix"] = MIX
    obj.rate_limiter_pattern_map_["qps"] = QPS

    return obj
}

func (mgr *RateLimiterManager) Initialize(rate_limiter_config_file string) {
    mgr.rate_limiter_ = NewTokenBucketRateLimiter()
    mgr.rate_limiter_config_ = NewRateLimiterConfig()
    mgr.rate_limiter_config_.Initialize(rate_limiter_config_file)
    mgr.rate_limiter_mix_pattern_     = mgr.rate_limiter_config_.rate_limiter_mix_pattern_
    mgr.rate_limiter_cpu_pattern_     = mgr.rate_limiter_config_.rate_limiter_cpu_pattern_
    mgr.rate_limiter_latency_pattern_ = mgr.rate_limiter_config_.rate_limiter_latency_pattern_
    mgr.rate_limiter_qps_pattern_     = mgr.rate_limiter_config_.rate_limiter_qps_pattern_
    mgr.update_qps_time_ms_           = int32(mgr.rate_limiter_config_.update_qps_time_ms_)
    // mgr.BuildRateLimiterOpsInfo()  // TODO 暂时只支持mix模式
    mgr.high_load_qps_number_ = int32(mgr.rate_limiter_mix_pattern_.high_load_qps_number_)

    //  TODO 公共变量赋值
    // auto pattern_param = GetRateLimiterPatternParam();  // std::map<std::string, double*>*
    // if (pattern_param->find("max_qps_permits") != pattern_param->end()) {
        // max_qps_curr_ = *((*pattern_param)["max_qps_permits"]);
    // }
    mgr.max_qps_curr_ = mgr.rate_limiter_mix_pattern_.max_qps_permits_  // TODO 暂时固定为mix pattern
    mgr.curr_sample_number_ = int32(mgr.rate_limiter_config_.max_sample_number_)
    mgr.request_latency_vec_ = make([]int64, int64(mgr.rate_limiter_config_.max_sample_number_))
    mgr.high_load_qps_vec_ = make([]int64, int64(mgr.rate_limiter_mix_pattern_.high_load_qps_number_))
    for i := range mgr.high_load_qps_vec_ {
        mgr.high_load_qps_vec_[i] = int64(mgr.max_qps_curr_)
    }
    // fmt.Printf("request_latency_vec_: %#v\n", mgr.request_latency_vec_)
    // fmt.Printf("high_load_qps_vec_: %#v\n", mgr.high_load_qps_vec_)

    // 混合模式初始化最大QPS
    mgr.high_load_qps_ = int32(mgr.rate_limiter_mix_pattern_.max_qps_permits_)
    //  开启线程
    if mgr.rate_limiter_config_.real_time_update_enable_ {
        // cpu_info_->Run()  // docker cpu 启动时自动执行
        mgr.rate_limiter_.Init(mgr.max_qps_curr_)
        mgr.rate_limiter_.SetRate(mgr.max_qps_curr_)
        // update_max_qps_thread_ = std::thread(&RateLimiterManager::UpdateMaxQps, this);
        go func() {
            mgr.UpdateMaxQps()  // TODO wait done
        }()
        log.Printf("rate limiter pattern: %s, max_qps_permits_: %f [rate limiter Started]\n",
                mgr.rate_limiter_config_.rate_limiter_type_, mgr.max_qps_curr_)
    }
}

func (mgr *RateLimiterManager) UpdateMaxQps() {
    for mgr.run_ {
        if !mgr.rate_limiter_config_.rate_limiter_enable_ {
            time.Sleep(time.Duration(mgr.update_qps_time_ms_) * time.Millisecond)
            continue
        }
        pattern := mgr.GetRateLimiterPattern()
        switch pattern {
            case CPU:
                // mgr.UpdateMaxQpsByCPU()
            case LATENCY:
                // mgr.UpdateMaxQpsByLatency()
            case QPS:
                // mgr.UpdateMaxQpsByQPS()
            case MIX:
                mgr.UpdateMaxQpsByMix()
            default:
                mgr.UpdateMaxQpsByMix()
        }  // switch
    }  // for
}

func (mgr *RateLimiterManager) GetRateLimiterPattern() int32 {
    return MIX  // TODO 暂时只支持mix模式
}

//  mix模式-根据cpu和延迟自适应
func (mgr *RateLimiterManager) UpdateMaxQpsByMix() {
    //  数组变量更新
    if mgr.high_load_qps_number_ != int32(mgr.rate_limiter_mix_pattern_.high_load_qps_number_) {
        mgr.high_load_qps_number_ = int32(mgr.rate_limiter_mix_pattern_.high_load_qps_number_)
        mgr.high_load_qps_vec_ = make([]int64, mgr.high_load_qps_number_)
        for i := range mgr.high_load_qps_vec_ {
            mgr.high_load_qps_vec_[i] = int64(mgr.max_qps_curr_)
        }
        mgr.high_load_qps_position_ = 0
    }

    // 统计事实QPS和处理QPS
    now_qps := atomic.LoadInt32(&mgr.current_qps_)
    now_qps_handle := atomic.LoadInt32(&mgr.current_qps_handle_)
    atomic.StoreInt32(&mgr.current_qps_, 0)
    atomic.StoreInt32(&mgr.current_qps_handle_, 0)

    var curr_latency int64 = 0
    var tp99_postion int32 = int32(mgr.rate_limiter_config_.max_sample_number_ * 0.99)
    var max_qps_curr int32 = int32(mgr.max_qps_curr_)
    var current_cpu_usage int32 = int32(GetCPURatio() * 100)

    if (mgr.rate_limiter_mix_pattern_.max_cpu_usage_ - mgr.rate_limiter_mix_pattern_.max_cpu_delta_ <= float64(current_cpu_usage)) &&
            (float64(current_cpu_usage) <= mgr.rate_limiter_mix_pattern_.max_cpu_usage_ + mgr.rate_limiter_mix_pattern_.max_cpu_delta_) {
        request_latency_vec := mgr.request_latency_vec_
        sort.Slice(request_latency_vec, func(i, j int) bool { return request_latency_vec[i] < request_latency_vec[j] })

        // 在高负载时记录4个qps值，取最大的。
        mgr.high_load_qps_count_++
        var high_load_qps int32 = 0
        if now_qps_handle <= 0 {
            high_load_qps = int32(mgr.max_qps_curr_)
        } else {
            high_load_qps = now_qps_handle
        }
        mgr.high_load_qps_position_ = int32(math.Mod(float64(mgr.high_load_qps_position_), mgr.rate_limiter_mix_pattern_.high_load_qps_number_))
        if mgr.high_load_qps_position_ < int32(len(mgr.high_load_qps_vec_)) {
            mgr.high_load_qps_vec_[mgr.high_load_qps_position_] = int64(high_load_qps)
        }
        if (float64(mgr.high_load_qps_count_) >= mgr.rate_limiter_mix_pattern_.high_load_qps_number_) {
            _, maxVal := MinMax(mgr.high_load_qps_vec_[0:int32(mgr.rate_limiter_mix_pattern_.high_load_qps_number_)])
            mgr.high_load_qps_ = int32(maxVal)
            mgr.high_load_latency_ = request_latency_vec[tp99_postion]
        }
        mgr.high_load_qps_position_++

    } else if float64(current_cpu_usage) <= mgr.rate_limiter_mix_pattern_.min_cpu_usage_ {
        sum := float64(SumOfArray(mgr.request_latency_vec_))
        if len(mgr.request_latency_vec_) >= 1 {
            mgr.low_load_latency_ = int64(sum / float64(len(mgr.request_latency_vec_)))
        }
        mgr.high_load_qps_vec_ = mgr.high_load_qps_vec_[:0]
        mgr.high_load_qps_position_ = 0
        mgr.high_load_qps_count_ = 0
    } else {
        mgr.high_load_qps_vec_ = mgr.high_load_qps_vec_[:0]
        mgr.high_load_qps_position_ = 0
        mgr.high_load_qps_count_ = 0
    }

    if current_cpu_usage >= int32(mgr.rate_limiter_mix_pattern_.max_cpu_usage_) {
        mgr.max_qps_curr_ = mgr.max_qps_curr_ * (1.0 - mgr.rate_limiter_mix_pattern_.adjust_load_percent_) +
                        float64(now_qps_handle) * mgr.rate_limiter_mix_pattern_.adjust_load_percent_ / 2.0
    } else if current_cpu_usage >= int32(mgr.rate_limiter_mix_pattern_.max_cpu_usage_ - mgr.rate_limiter_mix_pattern_.max_cpu_delta_) {
        mgr.max_qps_curr_ = mgr.max_qps_curr_ * (1.0 - mgr.rate_limiter_mix_pattern_.adjust_load_percent_) +
                        float64(now_qps_handle) * mgr.rate_limiter_mix_pattern_.adjust_load_percent_
    } else if current_cpu_usage <= int32(mgr.rate_limiter_mix_pattern_.min_cpu_usage_) {
        request_latency_vec := mgr.request_latency_vec_
        sort.Slice(request_latency_vec, func(i, j int) bool { return request_latency_vec[i] < request_latency_vec[j] })
        curr_latency = request_latency_vec[tp99_postion]
        if float64(now_qps_handle) > float64(mgr.high_load_qps_) * 0.9 {
            mgr.max_qps_curr_ = math.Min(
                float64(mgr.high_load_qps_) * 2, mgr.max_qps_curr_ * (1.0 + mgr.rate_limiter_mix_pattern_.adjust_load_percent_))
            mgr.high_load_qps_ = int32(mgr.max_qps_curr_)
        } else {
            mgr.max_qps_curr_ = math.Min(
                float64(mgr.high_load_qps_), mgr.max_qps_curr_ * (1.0 + mgr.rate_limiter_mix_pattern_.adjust_load_percent_))
        }
    } else if current_cpu_usage < int32(mgr.rate_limiter_mix_pattern_.max_cpu_usage_ - mgr.rate_limiter_mix_pattern_.max_cpu_delta_) {
        if now_qps > now_qps_handle {
            mgr.high_load_qps_ = int32(math.Min(float64(INT_MAX), float64(mgr.high_load_qps_) * (1.0 + mgr.rate_limiter_mix_pattern_.adjust_load_percent_)))
        }
        if mgr.high_load_latency_ != 0 {
            request_latency_vec := mgr.request_latency_vec_
            sort.Slice(request_latency_vec, func(i, j int) bool { return request_latency_vec[i] < request_latency_vec[j] })
            curr_latency = request_latency_vec[tp99_postion]
            if curr_latency < int64(float64(mgr.high_load_latency_) * (1.0 - mgr.rate_limiter_mix_pattern_.adjust_load_percent_)) {
                mgr.max_qps_curr_ = math.Min(
                    float64(mgr.high_load_qps_), mgr.max_qps_curr_ * (1.0 + mgr.rate_limiter_mix_pattern_.adjust_load_percent_))
            }
        } else {
            mgr.max_qps_curr_ = math.Min(
                float64(mgr.high_load_qps_), mgr.max_qps_curr_ * (1.0 + mgr.rate_limiter_mix_pattern_.adjust_load_percent_))
        }
    }

    mgr.max_qps_curr_ = math.Max(mgr.max_qps_curr_, mgr.rate_limiter_mix_pattern_.min_qps_permits_)
    mgr.rate_limiter_.SetRate(mgr.max_qps_curr_)

    // TODO LOG_EVERY_N
    log.Printf(`UpdateMaxQpsByMix: current_cpu_usage=%d, now_qps=%d, now_qps_handle=%d, high_load_qps=%d
                , max_cpu_usage=%0.2f, min_cpu_usage=%0.2f, high_load_latency=%d, low_load_latency=%d
                , curr_latency=%d, is_rate_limit=%t, before max_qps_curr=%d, after max_qps_curr=%0.2f`,
        current_cpu_usage, now_qps, now_qps_handle, mgr.high_load_qps_, mgr.rate_limiter_mix_pattern_.max_cpu_usage_,
        mgr.rate_limiter_mix_pattern_.min_cpu_usage_, mgr.high_load_latency_, mgr.low_load_latency_, curr_latency,
        mgr.is_rate_limit_, max_qps_curr, mgr.max_qps_curr_)

    time.Sleep(time.Duration(mgr.rate_limiter_config_.update_qps_time_ms_) * time.Millisecond)
}

func (mgr *RateLimiterManager) TryAquire() bool {
    if !mgr.rate_limiter_config_.rate_limiter_enable_ {
        return false
    }
    atomic.AddInt32(&mgr.current_qps_, 1)
    mgr.is_rate_limit_ = mgr.rate_limiter_.TryAquire(1, 0)
    if !mgr.is_rate_limit_ {
        atomic.AddInt32(&mgr.current_qps_handle_, 1)
    }
    return mgr.is_rate_limit_
}

func MinMax(array []int64) (int64, int64) {
    var max int64 = array[0]
    var min int64 = array[0]
    for _, value := range array {
        if max < value {
            max = value
        }
        if min > value {
            min = value
        }
    }
    return min, max
}

func SumOfArray(arr []int64) int64{
   var res int64 = 0
   for i := 0; i < len(arr); i++ {
      res += arr[i]
   }
   return res
}
