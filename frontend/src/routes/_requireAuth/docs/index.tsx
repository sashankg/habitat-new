import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { useForm } from "react-hook-form";

export const Route = createFileRoute("/_requireAuth/docs/")({
  async loader({ context }) {
    const did = context.authManager.handle;
    const response = await context.authManager.fetch(
      `/xrpc/com.habitat.listRecords?repo=${did}&collection=com.habitat.docs`,
    );
    const data: {
      records: {
        uri: string;
        cid: string;
        value: HabitatDoc;
      }[];
    } = await response?.json();
    return data;
  },
  staleTime: 1000 * 60 * 60,
  component() {
    const router = useRouter();
    const { records } = Route.useLoaderData();
    const { authManager } = Route.useRouteContext();
    const navigate = Route.useNavigate();
    const form = useForm();
    return (
      <>
        <form
          onSubmit={form.handleSubmit(async () => {
            const did = authManager.handle;
            const response = await authManager.fetch(
              `/xrpc/com.habitat.putRecord`,
              "POST",
              JSON.stringify({
                repo: did,
                collection: "com.habitat.docs",
                record: {
                  name: "Untitled",
                  blob: null,
                },
              }),
            );
            const { uri } = await response?.json();
            navigate({
              to: "/docs/$did/$rkey",
              params: {
                did: authManager.handle ?? "",
                rkey: uri.split("/").at(-1) ?? "",
              },
            });
            router.invalidate({ filter: (x) => x.pathname === "/docs/" });
          })}
        >
          <button type="submit" aria-busy={form.formState.isSubmitting}>
            New
          </button>
        </form>
        <table>
          <thead>
            <tr>
              <th>Name</th>
            </tr>
          </thead>
          <tbody>
            {records.map((doc) => (
              <tr key={doc.cid}>
                <td>
                  <Link
                    to="/docs/$did/$rkey"
                    params={{
                      did: authManager.handle ?? "",
                      rkey: doc.uri.split("/").at(-1) ?? "",
                    }}
                  >
                    {doc.value.name || doc.uri}
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </>
    );
  },
});
