import { newGqlClient, CreateAccount } from "./test-helpers";
import { assert, default as chai } from "chai"
import { gql, Client } from '@urql/core';


chai.config.truncateThreshold = 0; // 0 means "don't truncate unexpected value, ever".

describe("Session Actions", () => {

    it("Integration Test 1", async () => {
        let gqlClientUnauthorized = newGqlClient({})
        const loginParams = {
            "email": "President@Biden.com",
            "password": "whitehouse"
        }
        let loginResult = await Login(gqlClientUnauthorized, loginParams)
        assert.isUndefined(loginResult.error)

        let gqlAccountClient = newGqlClient({accountToken: loginResult.data.login})
        let createSessionResult = await CreateSession(gqlAccountClient)
        assert.isUndefined(createSessionResult.error)
        let sessionID = createSessionResult.data.createSession.activeSession

        // Join Voters
        let getVoterTokenResult = await GetVoterToken(gqlAccountClient)
        assert.isUndefined(getVoterTokenResult.error)

        let gqlAccountAndVoterClient = newGqlClient({accountToken: loginResult.data.login, voterToken: getVoterTokenResult.data.voterToken})
        let getVoterResult = await GetVoter(gqlAccountAndVoterClient, sessionID)
        assert.isUndefined(getVoterResult.error)

        let searchResult = await MusicSearch(gqlAccountAndVoterClient, sessionID, "The Jackie")
        assert.isUndefined(searchResult.error)
        let songToVoteFor = searchResult.data.musicSearch[0]

        let newSongAddition = {
            id:         songToVoteFor.id,
            title:      songToVoteFor.title,
            artist:     songToVoteFor.artist,
            image:      songToVoteFor.image,
            vote:       "UP",
            action:     "ADD"
        }
        let voteResult = await UpdateQueue(gqlAccountAndVoterClient, sessionID, newSongAddition)
        assert.isUndefined(voteResult.error)

        let sessionResult = await GetSession(gqlAccountAndVoterClient, sessionID)
        assert.isUndefined(sessionResult.error)
        assert.equal(songToVoteFor.id, sessionResult.data.sessionState.queue[0].simpleSong.id)

        let songUpvote = {
            id:         songToVoteFor.id,
            vote:       "UP",
            action:     "ADD"
        }
        voteResult = await UpdateQueue(gqlAccountAndVoterClient, sessionID, songUpvote)
        assert.isUndefined(voteResult.error)

        sessionResult = await GetSession(gqlAccountAndVoterClient, sessionID)
        assert.isUndefined(sessionResult.error)
        assert.equal(2, sessionResult.data.sessionState.queue[0].votes)

        let songDownvote = {
            id:         songToVoteFor.id,
            vote:       "DOWN",
            action:     "ADD"
        }
        voteResult = await UpdateQueue(gqlAccountAndVoterClient, sessionID, songDownvote)
        assert.isUndefined(voteResult.error)

        sessionResult = await GetSession(gqlAccountAndVoterClient, sessionID)
        assert.isUndefined(sessionResult.error)
        assert.equal(1, sessionResult.data.sessionState.queue[0].votes)

        let endSessionResult = await EndSession(gqlAccountAndVoterClient, sessionID)
        assert.isUndefined(endSessionResult.error)
    })

})

interface LoginParams {
    email: string,
    password: string
}
async function Login(gqlclient: Client, loginParams: LoginParams) {
    const LOGIN = gql`
        mutation login ($accountLogin: AccountLogin!) {
            login(accountLogin:$accountLogin)
        }`

    let result = await gqlclient.mutation(LOGIN, { accountLogin: loginParams })
    return result
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
    return result
}

async function MusicSearch(gqlclient: Client, sessionID: Number, query: String) {
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
    return result
}

async function GetVoterToken(gqlclient: Client) {
    const GET_VOTER_TOKEN = gql`
        query getVoterToken {
            voterToken
        }
    `

    let result = await gqlclient.query(GET_VOTER_TOKEN, {})
    return result
}

async function GetVoter(gqlclient: Client, sessionID: Number) {
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
    return result
}

interface SongUpdate {
    id:         String
    title?:     String
    artist?:    String
    image?:     String
    vote:       String
    action:     String
}
async function UpdateQueue(gqlclient: Client, sessionID: Number, songUpdate: SongUpdate) {
    const UPDATE_QUEUE = gql`
        mutation UpdateQueue($sessionID: Int!, $song: SongUpdate!) {
            updateQueue(sessionID: $sessionID, song: $song) {
                numberOfVoters
            }
        }
    `;

    let result = gqlclient.mutation(UPDATE_QUEUE, { sessionID: sessionID, song: songUpdate })
    return result
}

async function GetSession(gqlclient: Client, sessionID: Number) {
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
    return result
}

async function EndSession(gqlclient: Client, sessionID: Number) {
    const END_SESSION = gql`
        mutation endSession($sessionID: Int!){
            endSession(sessionID: $sessionID)
        }
    `

    let result = await gqlclient.mutation(END_SESSION, { sessionID: sessionID })
    return result
}