# wej-api
API for WeJ music service

# Objects
- Song:
    Song info (name, artist, album cover, song file?)
    Votes?
    When song gets added to a queue reset votes, however store total vote info in db when a queue is closed
    total number of plays
- Queue:
    list of songs in the queue
    unique id that's generated when a new queue object is created
    unique id can be used to modify the queue using api
    Currently playing song
- App:
    Stores all of the queues


# Business
Three use tiers:
- Individual
    Used for small gatherings/parties
    20-50 users per queue max
    free but show ads on the queue display
- "Bar"
    Used at bars or restaurants
    500-1000 users per queue max
    No ads, $10 per month? (should calculate an estimate of aws cost for a month of use)
- "Unlimited"
    Used at large events (sports games, concerts)
    Unlimited users per queue
    No ads, $100 per month? (again, should calculate an estimate. This will probably be expensive)
    Maybe do a DJ mode where songs are only played for 15-30 seconds to go through more songs
Could pitch by saving money by replacing a DJ