import { createFileRoute, Outlet, } from "@tanstack/react-router";

export const Route = createFileRoute('/_requireAuth/docs')({
  component() {
    return (
      <>
        <h1>Docs</h1>
        <Outlet />
      </>
    )
  }
});
