import { RiThumbDownLine, RiThumbUpLine } from "@remixicon/react"

import { Button } from "@/components/ui/button"

export function CanteenVoteButtons({
  likeCount,
  dislikeCount,
  myVote,
  disabled,
  onVote,
}: {
  likeCount: number
  dislikeCount: number
  myVote: number
  disabled?: boolean
  onVote: (voteType: number) => void
}) {
  return (
    <div className="flex items-center gap-2">
      <Button
        type="button"
        size="xs"
        variant={myVote === 1 ? "default" : "outline"}
        disabled={disabled}
        aria-pressed={myVote === 1}
        onClick={() => onVote(myVote === 1 ? 0 : 1)}
      >
        <RiThumbUpLine />
        {likeCount}
      </Button>
      <Button
        type="button"
        size="xs"
        variant={myVote === -1 ? "destructive" : "outline"}
        disabled={disabled}
        aria-pressed={myVote === -1}
        onClick={() => onVote(myVote === -1 ? 0 : -1)}
      >
        <RiThumbDownLine />
        {dislikeCount}
      </Button>
    </div>
  )
}
