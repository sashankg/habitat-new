import { useMutation } from "@tanstack/react-query";
import { createFileRoute, redirect } from "@tanstack/react-router";
import type { FormEvent } from "react";
import * as client from "openid-client";
import clientMetadata from "../../client-metadata";

const { client_id, redirect_uris } = clientMetadata(__DOMAIN__);
//const domain = "privi.dwelf-mirzam.ts.net"; //"habitat-new.onrender.com"
const domain = "habitat-new.onrender.com";
const config = new client.Configuration(
  {
    issuer: `https://${domain}/oauth/authorize`,
    authorization_endpoint: `https://${domain}/oauth/authorize`,
    token_endpoint: `https://${domain}/oauth/token`,
  },
  client_id || "",
);

export const Route = createFileRoute("/oauth-login")({
  validateSearch(search) {
    if (!search.code) {
      return {};
    }
    return {
      code: search.code as string,
    };
  },
  async beforeLoad({ search }) {
    const state = localStorage.getItem("state");
    if (search.code && state) {
      const token = await client.authorizationCodeGrant(
        config,
        new URL(window.location.href),
        {
          expectedState: state,
        },
      );

      console.log(token.access_token);
      return redirect({ to: "/" });
    }
  },
  component() {
    const { mutate: handleSubmit, isPending } = useMutation({
      async mutationFn(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const formData = new FormData(e.target as HTMLFormElement);
        const handle = formData.get("handle") as string;

        const state = client.randomState();
        localStorage.setItem("state", state);

        const url = client.buildAuthorizationUrl(config, {
          redirect_uri: redirect_uris[0],
          response_type: "code",
          handle,
          state: state,
        });

        window.location.href = url.toString();
      },
      onError(e) {
        console.error(e);
      },
    });
    return (
      <article>
        <h1>Login</h1>
        <form onSubmit={handleSubmit}>
          <input
            name="handle"
            type="text"
            placeholder="Handle"
            required
            defaultValue={"sashankg.bsky.social"}
          />
          <button aria-busy={isPending} type="submit">
            Login
          </button>
        </form>
      </article>
    );
  },
});
