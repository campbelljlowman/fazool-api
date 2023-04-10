import { gql, Client, cacheExchange, fetchExchange, subscriptionExchange } from '@urql/core';
import { createClient as createWSClient } from 'graphql-ws';
import { WebSocket}  from 'ws';

const gqlHTTPServerURL = "http://localhost:8080/query"
const gqlWSServerURL = "ws://localhost:8080/query"

const wsClient = createWSClient({
    url: gqlWSServerURL,
    webSocketImpl: WebSocket
});

interface AuthTokens {
    accountToken?: string,
    voterToken?: string
}

export function newGqlClient(authTokens: AuthTokens) {
    let gqlClient = new Client({
        url: gqlHTTPServerURL,
        exchanges: [
            cacheExchange, 
            fetchExchange,
            subscriptionExchange({
                forwardSubscription(request) {
                  const input = { ...request, query: request.query || '' }
                  return {
                    subscribe(sink) {
                      const unsubscribe = wsClient.subscribe(input, sink)
                      return { unsubscribe };
                    },
                  };
                },
              }),
            ],
        fetchOptions: () => {
          return {
            headers: {  AccountAuthentication: authTokens.accountToken ? `Bearer ${authTokens.accountToken}` : '' ,
                        VoterAuthentication: authTokens.voterToken ? `Bearer ${authTokens.voterToken}` : '' },
          };
        },
    })

    return gqlClient
}

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