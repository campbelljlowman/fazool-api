import { assert, default as chai } from "chai"
import { gql, Client, cacheExchange, fetchExchange } from '@urql/core';

chai.config.truncateThreshold = 0; // 0 means "don't truncate unexpected value, ever".

describe("Create and Remove account", () => {
    const gqlServerURL = "http://localhost:8080/query"

    const gqlClientUnauthorized = new Client({
        url: gqlServerURL,
        exchanges: [cacheExchange, fetchExchange],
    });

    it("Integration Test 1", async () => {
        let createAccountResult = await CreateAccount(gqlClientUnauthorized)
        assert.isUndefined(createAccountResult.error)

        let accountToken = createAccountResult.data.createAccount
        let gqlClientAuthorized = new Client({
            url: gqlServerURL,
            exchanges: [cacheExchange, fetchExchange],
            fetchOptions: () => {
              return {
                headers: { AccountAuthentication: accountToken ? `Bearer ${accountToken}` : '' },
              };
            },
        })

        let getAccountResult = await GetAccount(gqlClientAuthorized)
        assert.isUndefined(getAccountResult.error)

        let accountID = getAccountResult.data.account.id
        let deleteAccountResult = await DeleteAccount(gqlClientAuthorized, accountID)

        assert.isUndefined(deleteAccountResult.error)
    })
})

async function CreateAccount(gqlclient: Client) {
    const CREATE_ACCOUNT = gql`
        mutation createAccount ($newAccount: NewAccount!) {
            createAccount(newAccount: $newAccount)
        }
    `;

    const newAccount = {
        "firstName": "Anthony",
        "lastName": "Kedis",
        "email": "red@hot.chilipepperssw",
        "password": "cantstop"
    };

    let result = await gqlclient.mutation(CREATE_ACCOUNT, { newAccount: newAccount })
    return result
}

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
        mutation deleteAccount($accountID: Int!) {
            deleteAccount(accountID: $accountID)
        }`

    let result = await gqlclient.mutation(DELETE_ACCOUNT, { accountID: accountID})
    return result
}