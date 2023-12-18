# mq

One mq system based on Postgres table

# Todo

- Add doc
- Add test

# Increase consume speed

If current consume speed is not satisfying, try this alternative approach

```
func cron() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			go process()
		}
	}
}

func process() {
    const BatchProcessSize = 5
	capCh := make(chan struct{}, BatchProcessSize)

    // retrive limited number of messages to be consumed from table `queues`
    // query = select id from queues where is_dead = false and check_at < $1 limit 500
    // get queue ids to be consumed
    ids := []int64{}

	for _, id := range ids {
		capCh <- struct{}{}
		go func(queueID int64) {
			defer func() {
				<-capCh
			}()

            // get queue with lock from queue id, but skip locked
			c.consumeSingleMessage(queueID)
		}(id)
	}
}
```
