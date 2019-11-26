package main

import (
    "bytes"
    "bufio"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"
)

func test1() {
    cmd := exec.Command("tr", "a-z", "A-Z")
    cmd.Stdin = strings.NewReader("some input")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("in all caps: %q\n", out.String())
}

func run_with_env_variable() {
    cmd := exec.Command("prog")
    cmd.Env = append(os.Environ(),
        "FOO=duplicate_value", // ignored
        "FOO=actual_value",    // this value is used
    )
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }
}

func on_error() {
    cmd := exec.Command("ls", "/tmp/aaaaaa")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal("Run cmd error: ", err)
    }
    fmt.Printf("%s\n", out.String())
}

func realtime_output() {
    args := "-i test.mp4 -acodec copy -vcodec copy -f flv rtmp://aaa/bbb"
    cmd := exec.Command("ffmpeg", strings.Split(args, " ")...)

    stderr, _ := cmd.StderrPipe()
    cmd.Start()

    scanner := bufio.NewScanner(stderr)
    scanner.Split(bufio.ScanWords)
    for scanner.Scan() {
        m := scanner.Text()
        fmt.Println(m)
    }
    cmd.Wait()
}

func realtime_reader() {
    cmd := exec.Command("./test")

    output, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal("Open pipe error: ", err)
    }

    if err := cmd.Start(); err != nil {
        log.Fatal("Run cmd error: ", err)
    }

    scanner := bufio.NewScanner(output)
    for scanner.Scan() {
        fmt.Println(scanner.Text()) // Println will add back the final '\n'
    }
    if err := scanner.Err(); err != nil {
        log.Fatal("Reading pipe error: ", err)
    }

    if err := cmd.Wait(); err != nil {
        log.Fatal(err)
    }
}

/*
 * func main() {
 *     for i := 1; i <= 10; i++ {
 *         fmt.Printf("Running %d...\n", i)
 *         time.Sleep(time.Second * 1)
 *     }
 * }
 */
