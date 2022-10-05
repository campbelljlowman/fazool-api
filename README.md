# Fazool
API for Fazool music service

# Models
- Song:
    - Song info (name, artist, album cover, song file?)
    - Votes?
    - When song gets added to a queue reset votes, however store total vote info in db when a queue is closed
    - Total number of plays
    - Link/function to play
- Session:
    - list of songs in the queue
    - unique id that's generated when a new queue object is created
    - unique id can be used to modify the queue using api
    - Currently playing song - time started and whether it's paused
    - size
    - User who created the queue
    - voters who are allowed to vote (give out api tokens with a code scan, give them a timeout)
    - Explicit song filter
    - Streaming service to target
- App:
    - Start api, connect to db (and cache)
    - Stores all of the queues
- User
    - login creds
    - account level (Free, bar, unlimited)

# Technologies
- Backend language: GO
    Fast, multi platform, good way to learn
- Frontend language: React Native
    Multi platform, reactive
- API framework: Gin
    Lightweight and fast, don't need anything more than that
- Database: postgres
    available as a service, durable, well established
- Distributed cache: Redis
    open source, can store hashes, scales well

# Business
Three use tiers for facilitators:
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
    
Two revenue streams from users:
- Subscription
    Users could sign up to pay a small amount per month ($1-$3) to get perks
    Double votes
    Down votes
- Bonus votes  
    Users could purchase bonus votes which they could use on songs
    Votes cost $0.5-$1
    Discount for bulk purchases
    Session could control how many bonus votes can be used per person per song to limit takeovers
    Could give part of the bonus vote money used back to the establishment to incentivise them to use it
Could pitch by saving money by replacing a DJ

# User story
- As a Fazool facilitator, I can create a session of the app to display publicly that contains a currently playing song, a queue of songs that are up next with a number of votes, and a qr and number code for people to join and contribute
- As a Fazool user, I can join a session and see the queue of songs, add songs to the queue and vote for songs I want to hear

# Env setup

1. Install go 
2. cd to project root
3. run `go run .`
