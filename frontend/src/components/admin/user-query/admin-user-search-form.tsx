import { useState } from "react"
import { RiSearchLine } from "@remixicon/react"
import { EmailPrefixInput } from "@/components/auth/email-prefix-input"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { buildAuthEmail, normalizeAuthEmailPrefix } from "@/config/auth"
import type { FormSubmitEvent } from "../admin-utils"

type AdminUserQueryType = "email" | "username" | "review"

export interface AdminUserLookupFormValue {
  email?: string
  username?: string
  review_id?: number
}

interface AdminUserSearchFormProps {
  email: string
  username: string
  reviewID?: number
  emailDomain: string
  onSearch: (value: AdminUserLookupFormValue) => void
}

function getInitialQueryType(
  username: string,
  reviewID?: number
): AdminUserQueryType {
  if (username) return "username"
  if (reviewID) return "review"
  return "email"
}

export function AdminUserSearchForm({
  email,
  username,
  reviewID,
  emailDomain,
  onSearch,
}: AdminUserSearchFormProps) {
  const [queryType, setQueryType] = useState<AdminUserQueryType>(() =>
    getInitialQueryType(username, reviewID)
  )
  const [emailPrefix, setEmailPrefix] = useState(() =>
    normalizeAuthEmailPrefix(email, emailDomain)
  )
  const [usernameValue, setUsernameValue] = useState(username)
  const [reviewIDValue, setReviewIDValue] = useState(
    reviewID ? String(reviewID) : ""
  )

  function handleSubmit(event: FormSubmitEvent) {
    event.preventDefault()

    if (queryType === "email") {
      const nextPrefix = emailPrefix.trim()
      onSearch({
        email: nextPrefix
          ? buildAuthEmail(nextPrefix, emailDomain).toLowerCase()
          : undefined,
      })
      return
    }

    if (queryType === "username") {
      onSearch({ username: usernameValue.trim() || undefined })
      return
    }

    const nextReviewID = Number(reviewIDValue)
    onSearch({
      review_id:
        Number.isFinite(nextReviewID) && nextReviewID > 0
          ? Math.trunc(nextReviewID)
          : undefined,
    })
  }

  return (
    <form className="space-y-3" onSubmit={handleSubmit}>
      <Tabs
        value={queryType}
        onValueChange={(value) => setQueryType(value as AdminUserQueryType)}
      >
        <TabsList>
          <TabsTrigger value="email">邮箱</TabsTrigger>
          <TabsTrigger value="username">原始 username</TabsTrigger>
          <TabsTrigger value="review">点评 ID</TabsTrigger>
        </TabsList>
      </Tabs>

      <div className="flex max-w-md flex-wrap items-end gap-2">
        <div className="min-w-72 flex-1">
          {queryType === "email" ? (
            <EmailPrefixInput
              id="admin-user-email"
              label="邮箱"
              value={emailPrefix}
              emailDomain={emailDomain}
              onChange={setEmailPrefix}
              placeholder="jAccount"
            />
          ) : null}
          {queryType === "username" ? (
            <div className="space-y-2">
              <Label htmlFor="admin-user-username">原始 username</Label>
              <Input
                id="admin-user-username"
                value={usernameValue}
                onChange={(event) => setUsernameValue(event.target.value)}
                placeholder="username"
              />
            </div>
          ) : null}
          {queryType === "review" ? (
            <div className="space-y-2">
              <Label htmlFor="admin-user-review-id">点评 ID</Label>
              <Input
                id="admin-user-review-id"
                type="number"
                min={1}
                value={reviewIDValue}
                onChange={(event) => setReviewIDValue(event.target.value)}
                placeholder="review id"
              />
            </div>
          ) : null}
        </div>
        <Button type="submit">
          <RiSearchLine />
          查询
        </Button>
      </div>
    </form>
  )
}
