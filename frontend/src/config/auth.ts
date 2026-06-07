export const defaultAuthEmailDomain = "@sjtu.edu.cn"

export const authPasswordPolicy = {
  minLength: 10,
} as const

export function buildAuthEmail(
  prefix: string,
  domain = defaultAuthEmailDomain
) {
  return `${prefix.trim()}${domain}`
}

export function normalizeAuthEmailPrefix(
  value: string,
  domain = defaultAuthEmailDomain
) {
  const trimmed = value.trim()
  if (domain && trimmed.toLowerCase().endsWith(domain.toLowerCase())) {
    return trimmed.slice(0, -domain.length).split("@")[0]
  }
  return trimmed.split("@")[0]
}

export function validateAuthPassword(password: string) {
  if (/\s/.test(password)) {
    return "密码不能包含空白字符"
  }
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
