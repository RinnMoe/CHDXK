export const authEmailDomain = "@sjtu.edu.cn"

export const authPasswordPolicy = {
  minLength: 10,
} as const

export function buildAuthEmail(prefix: string) {
  return `${prefix.trim()}${authEmailDomain}`
}

export function normalizeAuthEmailPrefix(value: string) {
  return value.trim().replace(authEmailDomain, "").split("@")[0]
}

export function validateAuthPassword(password: string) {
  if (password.length < authPasswordPolicy.minLength) {
    return `密码至少 ${authPasswordPolicy.minLength} 位`
  }
  if (!/[A-Za-z]/.test(password)) {
    return "密码需要包含英文字母"
  }
  if (!/\d/.test(password)) {
    return "密码需要包含数字"
  }
  if (!/[^A-Za-z0-9]/.test(password)) {
    return "密码需要包含特殊字符"
  }
  return null
}
