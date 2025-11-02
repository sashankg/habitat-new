import clientMetadata from "..//client-metadata";
import * as client from "openid-client";

export class AuthManager {
  handle: string | null;
  private accessToken: string | null;
  private config: client.Configuration;

  constructor(serverDomain: string) {
    const { client_id } = clientMetadata(__DOMAIN__);
    this.config = new client.Configuration(
      {
        issuer: `https://${serverDomain}/oauth/authorize`,
        authorization_endpoint: `https://${serverDomain}/oauth/authorize`,
        token_endpoint: `https://${serverDomain}/oauth/token`,
      },
      client_id,
    );
    this.handle = localStorage.getItem("handle");
    this.accessToken = localStorage.getItem("token");
  }

  isAuthenticated() {
    return !!this.accessToken;
  }

  loginUrl(handle: string, redirectUri: string) {
    this.handle = handle;
    localStorage.setItem("handle", handle);
    const state = client.randomState();
    console.log(state);
    localStorage.setItem("state", state);
    return client.buildAuthorizationUrl(this.config, {
      redirect_uri: redirectUri,
      response_type: "code",
      handle,
      state,
    });
  }
  async exchangeCode(currentUrl: string) {
    const state = localStorage.getItem("state");
    if (!state) {
      throw new Error("No state found");
    }
    localStorage.removeItem("state");
    console.log(currentUrl);
    console.log(state);
    const token = await client.authorizationCodeGrant(
      this.config,
      new URL(currentUrl),
      {
        expectedState: state,
      },
    );
    this.accessToken = token.access_token;
    localStorage.setItem("token", token.access_token);
  }

  async fetch(
    url: string,
    method: string = "GET",
    body?: client.FetchBody,
    headers?: Headers,
    options?: client.DPoPOptions,
  ) {
    if (!this.accessToken) {
      throw new UnauthenticatedError();
    }
    if (!headers) {
      headers = new Headers();
    }
    headers.append("Habitat-Auth-Method", "oauth");
    const response = await client.fetchProtectedResource(
      this.config,
      this.accessToken,
      new URL(url),
      method,
      body,
      headers,
      options,
    );

    if (response.status === 401) {
      throw new UnauthenticatedError();
    }
    return response;
  }
}

export class UnauthenticatedError extends Error { }
