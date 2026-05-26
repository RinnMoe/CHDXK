const offlineReadCacheNames = ["read-api-cache"]

export async function clearOfflineReadCaches() {
  if (!("caches" in window)) return

  await Promise.all(
    offlineReadCacheNames.map((cacheName) => caches.delete(cacheName))
  )
}
