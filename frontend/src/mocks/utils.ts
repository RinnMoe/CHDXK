export function randomDelay() {
  const ms = Math.floor(Math.random() * 501)
  return new Promise((resolve) => setTimeout(resolve, ms))
}
