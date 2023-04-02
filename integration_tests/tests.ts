import { assert } from "chai"
import { gql, Client, cacheExchange, fetchExchange } from '@urql/core';

describe("Test GraphQL server", () => {
    const gqlServerURL = "http://localhost:8080/query"

    const gqlclient = new Client({
        url: gqlServerURL,
        exchanges: [cacheExchange, fetchExchange],
    });

    it("Register New User", () => {
        const CREATE_ACCOUNT = gql`
            mutation createAccount ($newAccount: NewAccount!) {
                createAccount(newAccount: $newAccount)
            }
        `;

        const newAccount = {
            "firstName": "Anthony",
            "lastName": "Kedis",
            "email": "red@hot.chilipeppers",
            "password": "cantstop"
        };

        gqlclient
        .mutation(CREATE_ACCOUNT, { newAccount: newAccount })
        .toPromise()
        .then(result => {
          console.log(result);
        });

    })

})
