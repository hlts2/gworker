# gworker

gworker is dispatch worker management library in golang

## Requirement
Go (>= 1.8)

## Installation

```shell
go get github.com/hlts2/gworker
```
## Example

```go

func main() {

    // Generate 3 workers.
    d := gworker.NewDispatcher(3)

    // Register finish monitoring of all jobs.
    // When all jobs are completed, a notification is sent to `<-d.Finish()`.
    d.StartJobObserver()

    for i := 0; i < 10; i++ {

        // Add job
        d.Add(func() error {
            time.Sleep(time.Second * 3)
            return nil
        })
    }

    // Starts workers
    d.Start()

    for i := 0; i < 100; i++ {
        d.Add(func() error {
            time.Sleep(time.Second * 2)
            return errors.New("errors.New")
        })
    }

    time.Sleep(time.Second * 10)

    // Scale up the number of workers
    d.UpScale(100)

FINISH_ALL_JOB:
    for {
        select {
        case err := <-d.JobError():  // get job error
            fmt.Println(err)
        case _ = <-d.Finish():       // When all the jobs are finished, notification comes
            break FINISH_ALL_JOB
        }
    }
}

```

## Author
[hlts2](https://github.com/hlts2)
