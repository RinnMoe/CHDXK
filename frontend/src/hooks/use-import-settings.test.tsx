import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderToStaticMarkup } from "react-dom/server"
import { describe, expect, it } from "vitest"
import { useImportSettings } from "./use-import-settings"

function Availability() {
  const methods = useImportSettings()
  return <span>{JSON.stringify(methods)}</span>
}

describe("import availability", () => {
  it("uses public switches independently, including disabling every entry", () => {
    const client = new QueryClient()
    client.setQueryData(
      ["system-settings", "public"],
      [
        { key: "import.school.enabled", value: "false" },
        { key: "import.manual.enabled", value: "true" },
        { key: "import.admin.enabled", value: "false" },
        { key: "import.mychu.visible", value: "false" },
      ]
    )
    const render = () =>
      renderToStaticMarkup(
        <QueryClientProvider client={client}>
          <Availability />
        </QueryClientProvider>
      ).replaceAll("&quot;", '"')
    expect(render()).toContain(
      '"school":false,"manual":true,"mychu":false,"admin":false'
    )
    client.setQueryData(
      ["system-settings", "public"],
      [
        { key: "import.school.enabled", value: "false" },
        { key: "import.manual.enabled", value: "false" },
        { key: "import.admin.enabled", value: "false" },
        { key: "import.mychu.visible", value: "false" },
      ]
    )
    expect(render()).not.toContain(":true")
    client.clear()
  })

  it("does not offer imports before settings load", () => {
    const client = new QueryClient()
    const html = renderToStaticMarkup(
      <QueryClientProvider client={client}>
        <Availability />
      </QueryClientProvider>
    )
    expect(html).not.toContain(":true")
    client.clear()
  })
})
