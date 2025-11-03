import { Editor, EditorContent, useEditor } from "@tiptap/react";
import { createFileRoute } from "@tanstack/react-router";
import StarterKit from "@tiptap/starter-kit";
import { useMemo } from "react";
import { useMutation } from "@tanstack/react-query";

export const Route = createFileRoute("/_requireAuth/docs/$did/$rkey")({
  async loader({ context, params }) {
    const { did, rkey } = params;
    const response = await context.authManager.fetch(
      `/xrpc/com.habitat.repo.getRecord?repo=${did}&collection=com.habitat.docs&rkey=${rkey}`,
    );

    const data: {
      uri: string;
      cid: string;
      value: HabitatDoc;
    } = await response?.json();

    return {
      ...data,
    };
  },
  component() {
    const { did, rkey } = Route.useParams();
    const { value } = Route.useLoaderData();
    const { authManager } = Route.useRouteContext();
    const { mutate: save } = useMutation({
      mutationFn: async ({ editor }: { editor: Editor }) => {
        const heading = editor.$node("heading")?.textContent;
        authManager.fetch(
          "/xrpc/com.habitat.repo.putRecord",
          "POST",
          JSON.stringify({
            repo: did,
            collection: "com.habitat.docs",
            rkey,
            record: {
              name: heading ?? "Untitled",
              blob: editor.getHTML(),
            },
          }),
        );
      },
    });
    // debounce
    const handleUpdate = useMemo(() => {
      let prevTimeout: number | undefined;
      return ({ editor }: { editor: Editor }) => {
        clearTimeout(prevTimeout);
        prevTimeout = setTimeout(() => {
          save({ editor });
        }, 1000);
      };
    }, [save]);
    const editor = useEditor({
      extensions: [StarterKit],
      content: value.blob || "",
      onUpdate: handleUpdate,
      editable: did === authManager.handle,
    });
    return (
      <article>
        <EditorContent editor={editor} />
      </article>
    );
  },
});
