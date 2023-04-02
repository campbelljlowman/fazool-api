"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@urql/core");
describe("Test GraphQL server", () => {
    const gqlServerURL = "http://localhost:8080/query";
    const gqlclient = new core_1.Client({
        url: gqlServerURL,
        exchanges: [core_1.cacheExchange, core_1.fetchExchange],
    });
    it("Register New User", () => {
        const CREATE_ACCOUNT = (0, core_1.gql) `
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
    });
});
