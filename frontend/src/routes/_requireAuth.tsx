import { UnauthenticatedError } from "@/auth";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_requireAuth")({
  beforeLoad({ context }) {
    if (!context.authManager.isAuthenticated()) {
      throw redirect({ to: "/oauth-login" });
    }
  },
  onCatch(error) {
    if (error instanceof UnauthenticatedError) {
      throw redirect({ to: "/oauth-login" });
    }
    throw error;
  },
  component() {
    return <Outlet />;
  },
});
