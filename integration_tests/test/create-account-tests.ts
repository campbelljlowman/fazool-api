import { assert, default as chai } from "chai"
import { gql, Client, cacheExchange, fetchExchange } from '@urql/core';

chai.config.truncateThreshold = 0; // 0 means "don't truncate unexpected value, ever".

describe("Test GraphQL server", () => {
    const gqlServerURL = "http://localhost:8080/query"

    const gqlclient = new Client({
        url: gqlServerURL,
        exchanges: [cacheExchange, fetchExchange],
    });

    it("Register New User", async () => {
        let data, error = await CreateAccount(gqlclient)
        assert.equal(error, undefined)
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

    try {
        let result = await gqlclient
        .mutation(CREATE_ACCOUNT, { newAccount: newAccount })
        .toPromise()
        console.log("Result: "+ result)
        return result.data, result.error

    } catch (error) {
        console.log("Error :" + error)
    }
    // .then(result => {
    //   console.log(result);
    // });

}
