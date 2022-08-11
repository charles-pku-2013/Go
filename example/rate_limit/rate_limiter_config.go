package system_metric

import (
    // "fmt"
    "log"
    "github.com/gurkankaymak/hocon"
)

type RateLimiterQPSPattern struct {
    max_qps_permits_        float64
}

type RateLimiterCPUPattern struct {
    min_qps_permits_        float64
    max_qps_permits_        float64
    //  一次调整负载增减的百分比
    adjust_load_percent_    float64
    max_cpu_usage_          float64
    min_cpu_usage_          float64
}

type RateLimiterMixPattern struct {
    min_qps_permits_         float64
    max_qps_permits_         float64
    //  一次调整负载增减的百分比
    adjust_load_percent_     float64
    max_cpu_usage_           float64
    max_cpu_delta_           float64
    min_cpu_usage_           float64
    //  采样区间的大小
    high_load_qps_number_    float64
}

type RateLimiterLatencyPattern struct {
    min_qps_permits_          float64
    max_qps_permits_          float64
    max_latency_ms_           float64
    min_latency_ms_           float64
    //  滑动窗口圈定请求数量
    min_sample_number_        float64
    change_sample_ratio_      float64
    //  一次调整负载增减的百分比
    adjust_load_percent_      float64
    //  在滑动窗口内，计算大于max_latency的请求比例
    max_latency_ratio_        float64
    //  在滑动窗口内，计算小于min_latency的请求比例
    min_latency_ratio_        float64
    //  在滑动窗口内，计算大于max_latency的请求比例下限
    max_latency_ratio_min_    float64
    //  QPS为0的时候，置空队列，恢复限流前状态
    empty_queue_ratio_        float64
}

type RateLimiterConfig struct {
    //  公用参数
    rate_limiter_type_               string
    //  启动限流功能
    rate_limiter_enable_             bool
    real_time_update_enable_         bool
    // 采样区间大小
    max_sample_number_               float64
    //  多长时间计算一次max_qps
    update_qps_time_ms_              float64
    //  限流模式
    rate_limiter_mix_pattern_        *RateLimiterMixPattern
    rate_limiter_cpu_pattern_        *RateLimiterCPUPattern
    rate_limiter_latency_pattern_    *RateLimiterLatencyPattern
    rate_limiter_qps_pattern_        *RateLimiterQPSPattern

    rate_limiter_conf_               hocon.Config
}

func NewRateLimiterConfig() *RateLimiterConfig {
    obj := new(RateLimiterConfig)
    obj.rate_limiter_type_ = "mix"
    obj.rate_limiter_enable_ = false
    obj.real_time_update_enable_ = false
    obj.max_sample_number_ = 1000.0
    obj.update_qps_time_ms_ = 1000.0

    obj.rate_limiter_mix_pattern_     = NewRateLimiterMixPattern()
    obj.rate_limiter_cpu_pattern_     = NewRateLimiterCPUPattern()
    obj.rate_limiter_latency_pattern_ = NewRateLimiterLatencyPattern()
    obj.rate_limiter_qps_pattern_     = NewRateLimiterQPSPattern()

    return obj
}

func NewRateLimiterQPSPattern() *RateLimiterQPSPattern {
    obj := new(RateLimiterQPSPattern)
    obj.max_qps_permits_ = 500.0
    return obj
}

func NewRateLimiterCPUPattern() *RateLimiterCPUPattern {
    obj := new(RateLimiterCPUPattern)
    obj.min_qps_permits_ = 50.0
    obj.max_qps_permits_ = 500.0
    obj.adjust_load_percent_ = 0.05
    obj.max_cpu_usage_ = 95.0
    obj.min_cpu_usage_ = 80.0
    return obj
}

func NewRateLimiterMixPattern() *RateLimiterMixPattern {
    obj := new(RateLimiterMixPattern)
    obj.min_qps_permits_ = 50.0
    obj.max_qps_permits_ = 500.0
    obj.adjust_load_percent_ = 0.03
    obj.max_cpu_usage_ = 96.0
    obj.max_cpu_delta_ = 3.0
    obj.min_cpu_usage_ = 80.0
    obj.high_load_qps_number_ = 4.0
    return obj
}

func NewRateLimiterLatencyPattern() *RateLimiterLatencyPattern {
    obj := new(RateLimiterLatencyPattern)
    obj.min_qps_permits_ = 50.0
    obj.max_qps_permits_ = 500.0
    obj.max_latency_ms_ = 600.0
    obj.min_latency_ms_ = 100.0
    obj.min_sample_number_ = 100.0
    obj.change_sample_ratio_ = 0.5
    obj.adjust_load_percent_ = 0.05
    obj.max_latency_ratio_ = 0.30
    obj.min_latency_ratio_ = 0.95
    obj.max_latency_ratio_min_ = 0.03
    obj.empty_queue_ratio_ = 0.2
    return obj
}

// NOTE!!! c必须用指针以返回修改
func (c *RateLimiterConfig) Initialize(conf_file string) {
    log.Println("RateLimiterConfig::Initialize()")  // DEBUG
    conf, err := hocon.ParseResource(conf_file)
    if err != nil {
        log.Fatal("RateLimiterConfig::Initialize read config error: ", err)
    }
    c.rate_limiter_conf_ = *conf
    // log.Println(c.rate_limiter_conf_)  // DEBUG

    // "rate_limiter"
    // NOTE!!! returns false "" 0.0 if not found
    rate_limiter_conf := c.rate_limiter_conf_.GetConfig("rate_limiter")
    if rate_limiter_conf == nil {
        log.Fatal("RateLimiterConfig::Initialize parse config error: rate_limiter is missing")
    }
    // log.Println(rate_limiter_conf)  // DEBUG
    c.rate_limiter_enable_ = rate_limiter_conf.GetBoolean("rate_limiter_enable")
    c.rate_limiter_type_ = rate_limiter_conf.GetString("rate_limiter_pattern")
    c.real_time_update_enable_ = rate_limiter_conf.GetBoolean("real_time_update_enable")
    c.max_sample_number_ = float64(rate_limiter_conf.GetInt("max_sample_number"))
    c.update_qps_time_ms_ = float64(rate_limiter_conf.GetInt("update_qps_time_ms"))

    // "cpu_pattern"
    cpu_pattern_conf := c.rate_limiter_conf_.GetConfig("cpu_pattern")
    if cpu_pattern_conf != nil {
        // log.Println(cpu_pattern_conf)  // DEBUG
        c.rate_limiter_cpu_pattern_.min_qps_permits_ = float64(cpu_pattern_conf.GetInt("min_qps_permits"))
        c.rate_limiter_cpu_pattern_.max_qps_permits_ = float64(cpu_pattern_conf.GetInt("max_qps_permits"))
        c.rate_limiter_cpu_pattern_.adjust_load_percent_ = cpu_pattern_conf.GetFloat64("adjust_load_percent")
        c.rate_limiter_cpu_pattern_.max_cpu_usage_ = float64(cpu_pattern_conf.GetInt("max_cpu_usage"))
        c.rate_limiter_cpu_pattern_.min_cpu_usage_ = float64(cpu_pattern_conf.GetInt("min_cpu_usage"))
    }

    // "mix_pattern"
    mix_pattern_conf := c.rate_limiter_conf_.GetConfig("mix_pattern")
    if mix_pattern_conf != nil {
        // log.Println(mix_pattern_conf)  // DEBUG
        c.rate_limiter_mix_pattern_.min_qps_permits_ = float64(mix_pattern_conf.GetInt("min_qps_permits"))
        c.rate_limiter_mix_pattern_.max_qps_permits_ = float64(mix_pattern_conf.GetInt("max_qps_permits"))
        c.rate_limiter_mix_pattern_.adjust_load_percent_ = mix_pattern_conf.GetFloat64("adjust_load_percent")
        c.rate_limiter_mix_pattern_.max_cpu_usage_ = float64(mix_pattern_conf.GetInt("max_cpu_usage"))
        c.rate_limiter_mix_pattern_.max_cpu_delta_ = float64(mix_pattern_conf.GetInt("max_cpu_delta"))
        c.rate_limiter_mix_pattern_.min_cpu_usage_ = float64(mix_pattern_conf.GetInt("min_cpu_usage"))
        c.rate_limiter_mix_pattern_.high_load_qps_number_ = float64(mix_pattern_conf.GetInt("high_load_qps_number"))
    }

    // "latency_pattern"
    latency_pattern_conf := c.rate_limiter_conf_.GetConfig("latency_pattern")
    if latency_pattern_conf != nil {
        // log.Println(latency_pattern_conf)  // DEBUG
        c.rate_limiter_latency_pattern_.min_qps_permits_ = float64(latency_pattern_conf.GetInt("min_qps_permits"))
        c.rate_limiter_latency_pattern_.max_qps_permits_ = float64(latency_pattern_conf.GetInt("max_qps_permits"))
        c.rate_limiter_latency_pattern_.max_latency_ms_ = float64(latency_pattern_conf.GetInt("max_latency_ms"))
        c.rate_limiter_latency_pattern_.min_latency_ms_ = float64(latency_pattern_conf.GetInt("min_latency_ms"))
        c.rate_limiter_latency_pattern_.min_sample_number_ = float64(latency_pattern_conf.GetInt("min_sample_number"))
        c.rate_limiter_latency_pattern_.change_sample_ratio_ = latency_pattern_conf.GetFloat64("change_sample_ratio")
        c.rate_limiter_latency_pattern_.adjust_load_percent_ = latency_pattern_conf.GetFloat64("adjust_load_percent")
        c.rate_limiter_latency_pattern_.max_latency_ratio_ = latency_pattern_conf.GetFloat64("max_latency_ratio")
        c.rate_limiter_latency_pattern_.min_latency_ratio_ = latency_pattern_conf.GetFloat64("min_latency_ratio")
        c.rate_limiter_latency_pattern_.max_latency_ratio_min_ = latency_pattern_conf.GetFloat64("max_latency_ratio_min")
        c.rate_limiter_latency_pattern_.empty_queue_ratio_ = latency_pattern_conf.GetFloat64("empty_queue_ratio")
    }

    // "qps_pattern"
    qps_pattern_conf := c.rate_limiter_conf_.GetConfig("qps_pattern")
    if qps_pattern_conf != nil {
        // log.Println(qps_pattern_conf)  // DEBUG
        c.rate_limiter_qps_pattern_.max_qps_permits_ = float64(qps_pattern_conf.GetInt("max_qps_permits"))
    }
}
