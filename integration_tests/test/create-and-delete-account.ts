import { assert, default as chai } from "chai"
import { gql, Client } from '@urql/core';
import { newGqlClient, CreateAccount } from "./test-helpers";

chai.config.truncateThreshold = 0; // 0 means "don't truncate unexpected value, ever".

describe("Create and Delete account", () => {

    it("Integration Test 1", async () => {
        let gqlClientUnauthorized = newGqlClient({})

        const newAccount = {
            "firstName": "Anthony",
            "lastName": "Kedis",
            "email": "red@hot.chilipeppers",
            "password": "cantstop"
        };
        let createAccountResult = await CreateAccount(gqlClientUnauthorized, newAccount)
        assert.isUndefined(createAccountResult.error)
        let accountToken = createAccountResult.data.createAccount

        let gqlClientAuthorized = newGqlClient({accountToken: accountToken})
        let getAccountResult = await GetAccount(gqlClientAuthorized)
        assert.isUndefined(getAccountResult.error)
        let accountID = getAccountResult.data.account.id

        let deleteAccountResult = await DeleteAccount(gqlClientAuthorized, accountID)
        assert.isUndefined(deleteAccountResult.error)
    })
})

async function GetAccount(gqlclient: Client) {
    const GET_ACCOUNT = gql`
        query getAccount {
            account {
                id
                firstName
                activeSession
            }
        }`

    let result = await gqlclient.query(GET_ACCOUNT, {})
    return result
}

async function DeleteAccount(gqlclient: Client, accountID: Number) {
    const DELETE_ACCOUNT = gql`
        mutation deleteAccount($targetAccountID: Int!) {
            deleteAccount(targetAccountID: $targetAccountID)
        }`

    let result = await gqlclient.mutation(DELETE_ACCOUNT, { targetAccountID: accountID})
    return result
}