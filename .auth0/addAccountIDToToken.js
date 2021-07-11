const axios = require("axios");

/**
 * Handler that will be called during the execution of a PostLogin flow.
 *
 * @param {Event} event - Details about the user and the context in which they are logging in.
 * @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
 */
exports.onExecutePostLogin = async (event, api) => {
    const user_email = event.user.identities;
    if (!user_email) {
        api.access.deny("could not determine user email address");
    }

    try {
        const res = await axios.post(
            `${event.secrets.endpoint_uri}/external/auth0/v1/login`,
            {
                auth0Subject: event.user.user_id,
                email: event.user.email,
                name: event.user.name,
            },
            {
                headers: {
                    "X-Auth0-ServerToken": event.secrets.server_token,
                },
            },
        );

        api.accessToken.setCustomClaim("https://accountID", res.data.id);
    } catch (err) {
        console.error({ message: err.response.data.message, statusCode: err.response.statusCode });
        api.access.deny("unable to procure user account id");
    }
};
