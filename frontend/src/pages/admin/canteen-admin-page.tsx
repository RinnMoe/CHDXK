import { useState } from "react"

import { CanteenDishPanel } from "@/components/admin/canteen/canteen-dish-panel"
import { CanteenHallPanel } from "@/components/admin/canteen/canteen-hall-panel"
import { CanteenMediaPanel } from "@/components/admin/canteen/canteen-media-panel"
import { CanteenReportPanel } from "@/components/admin/canteen/canteen-report-panel"
import { CanteenStorePanel } from "@/components/admin/canteen/canteen-store-panel"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Button } from "@/components/ui/button"

type CanteenAdminTab = "halls" | "stores" | "dishes" | "media" | "reports"

const tabs: { key: CanteenAdminTab; label: string }[] = [
  { key: "halls", label: "食堂" },
  { key: "stores", label: "店铺" },
  { key: "dishes", label: "菜品" },
  { key: "media", label: "实拍审核" },
  { key: "reports", label: "举报" },
]

export function CanteenAdminPage() {
  const [tab, setTab] = useState<CanteenAdminTab>("halls")

  return (
    <>
      <PageTitle>食堂管理</PageTitle>
      <PageShell>
        <div className="space-y-6">
          <header className="space-y-2">
            <h1 className="text-2xl font-bold">食堂管理</h1>
            <p className="text-sm text-muted-foreground">
              维护校区食堂、店铺目录、菜品与实拍审核。用户端只在开启「食堂评价」后可见。
            </p>
          </header>

          <div className="flex flex-wrap gap-2">
            {tabs.map((item) => (
              <Button
                key={item.key}
                type="button"
                variant={tab === item.key ? "default" : "outline"}
                size="sm"
                onClick={() => setTab(item.key)}
              >
                {item.label}
              </Button>
            ))}
          </div>

          {tab === "halls" && <CanteenHallPanel />}
          {tab === "stores" && <CanteenStorePanel />}
          {tab === "dishes" && <CanteenDishPanel />}
          {tab === "media" && <CanteenMediaPanel />}
          {tab === "reports" && <CanteenReportPanel />}
        </div>
      </PageShell>
    </>
  )
}
