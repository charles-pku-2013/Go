// NOTE now 是epoch到现在的时间戳单位是微秒

package system_metric

import (
    // "github.com/alibaba/sentinel-golang/util"
    // "io/ioutil"
    // "os"
    "log"
    "math"
    // "strconv"
    // "strings"
    "sync"
    // "sync/atomic"
    "time"
)

const MAX_QPS = 500

type TokenBucketRateLimiter struct {
    max_permits_    float64  // 最大令牌桶容量
    stored_permits_ float64  // 当前令牌桶内的令牌数
    next_free_      uint64   // 生成令牌的开始时间
    interval_       float64  // 生成令牌的速率，即限速的速率值，单位：秒/个
    mut_            sync.Mutex
}

func NewTokenBucketRateLimiter() *TokenBucketRateLimiter {
    obj := new(TokenBucketRateLimiter)
    obj.max_permits_    = MAX_QPS
    obj.stored_permits_ = 0.0
    obj.next_free_      = 0
    obj.interval_       = 1000000.0 / MAX_QPS
    return obj
}

func (limiter *TokenBucketRateLimiter) Init(max_permits float64) {
    if max_permits <= 0 {
        limiter.max_permits_ = MAX_QPS // 500
    } else {
        limiter.max_permits_ = max_permits
    }
    limiter.stored_permits_ = limiter.max_permits_
    limiter.interval_ = 1000000.0 / max_permits // microseconds
    log.Printf("TokenBucketLimiter Init max_permits=%0.2f", limiter.max_permits_)
}

// timeout_us 单位是毫秒
func (limiter *TokenBucketRateLimiter) TryAquire(permits float64, timeout_us uint64) bool {
    limiter.mut_.Lock()
    defer limiter.mut_.Unlock()
    now := uint64(time.Now().UnixNano() / 1000)  // timestamp in microseconds since epoch

    if limiter.next_free_ > now + timeout_us * 1000 {
        return true // 发生限流，丢弃请求
    } else {
        limiter.Aquire(permits, now) // 等待 timout_us 时长
    }

    return false // 未发生限流 放行请求
}

func (limiter *TokenBucketRateLimiter) Aquire(permits float64, now uint64) uint64 {
    if (permits <= 0) {
        log.Fatalln("TokenBucketLimiter Acquire permits must be greater than 0")
    }

    wait_time := limiter.ClaimNext(permits, now)  // microseconds
    time.Sleep(time.Duration(wait_time) * time.Microsecond)

    return wait_time / 1000.0  // return milliseconds
}

// returns microseconds
func (limiter *TokenBucketRateLimiter) ClaimNext(permits float64, now uint64) uint64 {
    limiter.Sync(now)

    wait := limiter.next_free_ - now
    stored := math.Min(permits, limiter.stored_permits_)  // 实际能申请到的令牌数量
    fresh := permits - stored   // 期望申请令牌数量与实际能申请到的数量之差,如果都能申请到则fresh=0
                                // fresh > 0 不能全部申请到
    next_free := uint64(fresh * limiter.interval_)
    limiter.next_free_ += next_free
    limiter.stored_permits_ -= stored  // 更新令牌库存

    return wait  // 返回微秒单位
}

func (limiter *TokenBucketRateLimiter) Sync(now uint64) {
    if now > limiter.next_free_ {
        step := float64(now - limiter.next_free_) / limiter.interval_
        limiter.stored_permits_ = math.Min(limiter.max_permits_, limiter.stored_permits_ + step)
        limiter.next_free_ = now
    }
}

func (limiter *TokenBucketRateLimiter) SetRate(rate float64) {
    if rate <= 0.0 {
        log.Println("TokenBucketLimiter SetRate rate must be greater than 0")
        return
    }

    limiter.mut_.Lock()
    defer limiter.mut_.Unlock()
    limiter.interval_ = 1000000.0 / rate
    limiter.max_permits_ = rate
    log.Printf("TokenBucketLimiter setting rate to %0.2f", limiter.max_permits_)
}

func (limiter *TokenBucketRateLimiter) GetRate() float64 {
    return limiter.max_permits_
}

// var (
    // max_permits_    float64 = MAX_QPS  // 最大令牌桶容量
    // stored_permits_ float64 = 0.0      // 当前令牌桶内的令牌数
    // next_free_      uint64  = 0        // 生成令牌的开始时间
    // interval_       float64 = 1000000.0 / MAX_QPS  // 生成令牌的速率，即限速的速率值，单位：秒/个
    // mut_            sync.Mutex
// )

// func RateLimiterInit(max_permits float64) {
    // if max_permits <= 0 {
        // max_permits_ = MAX_QPS // 500
    // } else {
        // max_permits_ = max_permits
    // }
    // stored_permits_ = max_permits_
    // interval_ = 1000000.0 / max_permits // microseconds
    // log.Printf("TokenBucketLimiter Init max_permits=%0.2f", max_permits_)
// }

// timeout_us 单位是毫秒
// func RateLimiterTryAquire(permits float64, timeout_us uint64) bool {
    // mut_.Lock()
    // defer mut_.Unlock()
    // now := uint64(time.Now().UnixNano() / 1000)  // timestamp in microseconds since epoch

    // if next_free_ > now + timeout_us * 1000 {
        // return true // 发生限流，丢弃请求
    // } else {
        // Aquire(permits, now) // 等待 timout_us 时长
    // }

    // return false // 未发生限流 放行请求
// }

// func Aquire(permits float64, now uint64) uint64 {
    // if (permits <= 0) {
        // log.Fatalln("TokenBucketLimiter Acquire permits must be greater than 0")
    // }

    // wait_time := ClaimNext(permits, now)  // microseconds
    // time.Sleep(time.Duration(wait_time) * time.Microsecond)

    // return wait_time / 1000.0  // return milliseconds
// }

// returns microseconds
// func ClaimNext(permits float64, now uint64) uint64 {
    // Sync(now)

    // wait := next_free_ - now
    // stored := math.Min(permits, stored_permits_)  // 实际能申请到的令牌数量
    // fresh := permits - stored   // 期望申请令牌数量与实际能申请到的数量之差,如果都能申请到则fresh=0
                                // fresh > 0 不能全部申请到
    // next_free := uint64(fresh * interval_)
    // next_free_ += next_free
    // stored_permits_ -= stored  // 更新令牌库存

    // return wait  // 返回微秒单位
// }

// func Sync(now uint64) {
    // if now > next_free_ {
        // step := float64(now - next_free_) / interval_
        // stored_permits_ = math.Min(max_permits_, stored_permits_ + step)
        // next_free_ = now
    // }
// }

// func RateLimiterSetRate(rate float64) {
    // if rate <= 0.0 {
        // log.Println("TokenBucketLimiter SetRate rate must be greater than 0")
        // return
    // }

    // mut_.Lock()
    // defer mut_.Unlock()
    // interval_ = 1000000.0 / rate
    // max_permits_ = rate
    // log.Printf("TokenBucketLimiter setting rate to %0.2f", max_permits_)
// }

// func RateLimiterGetRate() float64 {
    // return max_permits_
// }
