# gworker

gworker is dispatch worker management library written in golang.

## Requirement
Go (>= 1.8)

## Installation

```shell
go get github.com/hlts2/gworker
```
## Example

```go

func main() {

    // Generates 3 workers
    d := gworker.NewDispatcher(3)

    // Register finish monitoring of all jobs.
    // When all jobs are completed, a notification is sent to `<-d.Finish()`.
    d.StartJobObserver()

    for i := 0; i < 100; i++ {

        // Add adds job
        d.Add(func() error {
            time.Sleep(time.Second * 1)
            return nil
        })
    }

    // Start starts workers.
    d.Start()

    for i := 0; i < 100; i++ {
        d.Add(func() error {
            time.Sleep(time.Second * 2)
            return errors.New("errors.New")
        })
    }

    // Scale up the number of worker.
    d.UpScale(20)

FINISH_ALL_JOB:
    for {
        select {
        case err := <-d.JobError():  // get job error
            fmt.Println(err)
        case _ = <-d.Finish():       // When all the jobs are finished, notification comes
            break FINISH_ALL_JOB
        }
    }

    // Stop stops workers.
    // In the case of this example, 23 workers are stopped.
    d.Stop()
}

```

## Future
- [ ] Auto scaling of worker

Automatically scaled up when there are many jobs and fewer workers, while scaled down when there are fewer jobs and more workers.

## Author
[hlts2](https://github.com/hlts2)

## LICENSE
gworker released under MIT license, refer [LICENSE](https://github.com/hlts2/gworker/blob/master/LICENSE) file.

