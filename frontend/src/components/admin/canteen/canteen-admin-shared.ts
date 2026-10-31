import type { CanteenHallInput, CanteenStoreInput } from "@/api/canteen"

export const selectClass =
  "h-9 rounded-md border border-input bg-transparent px-2.5 text-sm shadow-xs outline-none focus-visible:ring-3 focus-visible:ring-ring/50"

export const emptyHall: CanteenHallInput = {
  campus_key: "benbu",
  name: "",
  short_name: "",
  address: "",
  status: "active",
  display_order: 0,
}

export const emptyStore: CanteenStoreInput = {
  hall_id: 0,
  name: "",
  floor: "",
  categories: [],
  status: "pending",
}
