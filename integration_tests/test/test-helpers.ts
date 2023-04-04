import { gql, Client, cacheExchange, fetchExchange } from '@urql/core';

export const gqlServerURL = "http://localhost:8080/query"

export const gqlClientUnauthorized = new Client({
    url: gqlServerURL,
    exchanges: [cacheExchange, fetchExchange],
});

interface newAccount {
    firstName: string
    lastName: string
    email: string
    password: string
}

export async function CreateAccount(gqlclient: Client, newAccount: newAccount) {
    const CREATE_ACCOUNT = gql`
        mutation createAccount ($newAccount: NewAccount!) {
            createAccount(newAccount: $newAccount)
        }
    `;

    let result = await gqlclient.mutation(CREATE_ACCOUNT, { newAccount: newAccount })
    return result
}