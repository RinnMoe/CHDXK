import { describe,it,expect } from "vitest"
import { normalizeAuthUser } from "./auth"
import { canAccessAdminTab } from "@/components/admin/admin-navigation-model"

describe("本站身份与分项管理权限",()=>{
 const account={id:8,username:"local",email:"",role:"admin",permissions:["users"]}
 it("管理员默认内容运维，不默认拥有超管功能",()=>{
  const user=normalizeAuthUser(account)
  expect(user.is_admin()).toBe(true)
  expect(user.can("users")).toBe(true)
  expect(user.can("settings")).toBe(false)
  expect(canAccessAdminTab("users",user.can)).toBe(true)
  expect(canAccessAdminTab("appearance",user.can)).toBe(false)
  expect(canAccessAdminTab("system-api-key",user.can)).toBe(false)
 })
 it("普通用户即使有遗留授权也不能使用管理功能",()=>{
  const user=normalizeAuthUser({...account,role:"user"})
  expect(user.can("users")).toBe(false)
 })
 it("超管拥有全部分项功能",()=>{
  const user=normalizeAuthUser({...account,role:"super_admin",permissions:[]})
  expect(user.is_super_admin()).toBe(true)
  for(const p of ["users","roles","points","settings","maintenance"]) expect(user.can(p)).toBe(true)
 })
})
