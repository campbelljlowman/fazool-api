import { newGqlClient } from "./test-helpers";
import { assert, default as chai } from "chai"
import { gql, Client } from '@urql/core';


chai.config.truncateThreshold = 0; // 0 means "don't truncate unexpected value, ever".

describe("Session Actions", () => {

    it("Integration Test 1", async () => {
        const adminLoginParams = {
            "email": "President@Biden.com",
            "password": "whitehouse"
        }
        let gqlAdminClient = await GetGqlClientForUser(adminLoginParams)
        let createSessionResult = await CreateSession(gqlAdminClient)

        let sessionID = createSessionResult.createSession.activeSession
        
        await RunSessionActions(gqlAdminClient, sessionID, "ADMIN")

        const privilegedVoterLoginParams = {
            "email": "mikey@gmail.com",
            "password": "gobraves"
        }
        let gqlPrivelegedVoterClient = await GetGqlClientForUser(privilegedVoterLoginParams)
        await RunSessionActions(gqlPrivelegedVoterClient, sessionID, "PRIVILEGED")
        
        let gqlFreeVoterClient = await GetGqlClientForUser()
        await RunSessionActions(gqlFreeVoterClient, sessionID, "FREE")

        await EndSession(gqlAdminClient, sessionID)
    })

})

async function GetGqlClientForUser(loginParams?: LoginParams) {
    let gqlClient = newGqlClient({})
    let accountToken

    if (loginParams) {
        let loginResult = await Login(gqlClient, loginParams)
        gqlClient = newGqlClient({accountToken: loginResult.login})
        accountToken = loginResult.login
    } else {
        accountToken = ''
    }

    let getVoterTokenResult = await GetVoterToken(gqlClient)

    gqlClient = newGqlClient({accountToken: accountToken, voterToken: getVoterTokenResult.voterToken})

    return gqlClient  
}

async function RunSessionActions(gqlclient: Client, sessionID: number, voterType: String) {
        await GetVoter(gqlclient, sessionID)

        let searchResult = await MusicSearch(gqlclient, sessionID, "The Jackie")
        let songToVoteFor = searchResult.musicSearch[0]

        // Add song to queue
        let newSongAddition = {
            id:         songToVoteFor.id,
            title:      songToVoteFor.title,
            artist:     songToVoteFor.artist,
            image:      songToVoteFor.image,
            vote:       "UP",
            action:     "ADD"
        }
        await UpdateQueue(gqlclient, sessionID, newSongAddition)

        let sessionResult = await GetSession(gqlclient, sessionID)
        assert.equal(songToVoteFor.id, sessionResult.sessionState.queue[0].simpleSong.id)

        if (voterType === "ADMIN") {
            let currentVotes = sessionResult.sessionState.queue[0].votes
            // Admin can vote unlimited
            let songUpvote = {
                id:         songToVoteFor.id,
                vote:       "UP",
                action:     "ADD"
            }
            await UpdateQueue(gqlclient, sessionID, songUpvote)
            sessionResult = await GetSession(gqlclient, sessionID)

            assert.equal(currentVotes + 1, sessionResult.sessionState.queue[0].votes)

            let songDownvote = {
                id:         songToVoteFor.id,
                vote:       "DOWN",
                action:     "ADD"
            }
            await UpdateQueue(gqlclient, sessionID, songDownvote)
            sessionResult = await GetSession(gqlclient, sessionID)

            assert.equal(currentVotes, sessionResult.sessionState.queue[0].votes)
        }

        if (voterType === "PRIVILEGED") {
            let currentVotes = sessionResult.sessionState.queue[0].votes
            // Bonus vote
            let songUpvote = {
                id:         songToVoteFor.id,
                vote:       "UP",
                action:     "ADD"
            }
            await UpdateQueue(gqlclient, sessionID, songUpvote)
            sessionResult = await GetSession(gqlclient, sessionID)

            assert.equal(currentVotes + 1, sessionResult.sessionState.queue[0].votes)
        }

        if(voterType === "FREE") {
            // Remove Upvote
            let currentVotes = sessionResult.sessionState.queue[0].votes
            let songDownvote = {
                id:         songToVoteFor.id,
                vote:       "UP",
                action:     "REMOVE"
            }
            await UpdateQueue(gqlclient, sessionID, songDownvote)
            sessionResult = await GetSession(gqlclient, sessionID)

            assert.equal(currentVotes - 1, sessionResult.sessionState.queue[0].votes)
        }
}

interface LoginParams {
    email: String,
    password: String
}
async function Login(gqlclient: Client, loginParams: LoginParams) {
    const LOGIN = gql`
        mutation login ($accountLogin: AccountLogin!) {
            login(accountLogin:$accountLogin)
        }`

    let result = await gqlclient.mutation(LOGIN, { accountLogin: loginParams })
    assert.isUndefined(result.error)
    return result.data
}

async function CreateSession(gqlclient: Client) {
    const CREATE_SESSION = gql`
        mutation createSession {
            createSession{
                activeSession
            }
        }
    `

    let result = await gqlclient.mutation(CREATE_SESSION, {})
    assert.isUndefined(result.error)
    return result.data
}

async function MusicSearch(gqlclient: Client, sessionID: number, query: String) {
    const MUSIC_SEARCH = gql`
        query musicSearch ($sessionID: Int!, $query: String!){
            musicSearch (sessionID: $sessionID, query: $query){
                id
                title
                artist
                image
            }
        }
    `

    let result = await gqlclient.query(MUSIC_SEARCH, { sessionID: sessionID, query: query })
    assert.isUndefined(result.error)
    return result.data
}

async function GetVoterToken(gqlclient: Client) {
    const GET_VOTER_TOKEN = gql`
        mutation getVoterToken {
            voterToken
        }
    `

    let result = await gqlclient.mutation(GET_VOTER_TOKEN, {})
    assert.isUndefined(result.error)
    return result.data
}

async function GetVoter(gqlclient: Client, sessionID: number) {
    const GET_VOTER = gql`
        query voter ($sessionID: Int!){
            voter (sessionID: $sessionID){
                type
                songsUpVoted
                songsDownVoted
                bonusVotes
            }
        }
    `
    
    let result = await gqlclient.query(GET_VOTER, { sessionID: sessionID })
    assert.isUndefined(result.error)
}

interface SongUpdate {
    id:         String
    title?:     String
    artist?:    String
    image?:     String
    vote:       String
    action:     String
}
async function UpdateQueue(gqlclient: Client, sessionID: number, songUpdate: SongUpdate) {
    const UPDATE_QUEUE = gql`
        mutation UpdateQueue($sessionID: Int!, $song: SongUpdate!) {
            updateQueue(sessionID: $sessionID, song: $song) {
                numberOfVoters
            }
        }
    `;

    let result = await gqlclient.mutation(UPDATE_QUEUE, { sessionID: sessionID, song: songUpdate })
    assert.isUndefined(result.error)
    return result.data
}

async function GetSession(gqlclient: Client, sessionID: number) {
    const GET_SESSION_STATE = gql`
        query getSessionState($sessionID: Int!){
            sessionState(sessionID: $sessionID){
                currentlyPlaying {
                    simpleSong{
                        id
                        title
                        artist
                        image
                    }
                    playing
                }
                queue {
                    simpleSong {
                        id
                        title
                        artist
                        image
                    }
                    votes
                }
                numberOfVoters
            }
        }
    `;

    let result = await gqlclient.query(GET_SESSION_STATE, { sessionID: sessionID })
    assert.isUndefined(result.error)
    return result.data
}

function SubscribeSessionState(gqlclient: Client, sessionID: number) {
    const SUBSCRIBE_SESSION_STATE = gql`
        subscription subscribeSessionState($sessionID: Int!){
            subscribeSessionState(sessionID: $sessionID){
                currentlyPlaying {
                    simpleSong{
                        id
                        title
                        artist
                        image
                    }
                    playing
                }
                queue {
                    simpleSong {
                        id
                        title
                        artist
                        image
                    }
                    votes
                }
                numberOfVoters
            }
        }
    `;

    
    const { unsubscribe } = gqlclient.subscription(SUBSCRIBE_SESSION_STATE, { sessionID: sessionID}).subscribe(result => {
        assert.isUndefined(result.error)
    });

    unsubscribe()
}

async function EndSession(gqlclient: Client, sessionID: number) {
    const END_SESSION = gql`
        mutation endSession($sessionID: Int!){
            endSession(sessionID: $sessionID)
        }
    `

    let result = await gqlclient.mutation(END_SESSION, { sessionID: sessionID })
    assert.isUndefined(result.error)
}