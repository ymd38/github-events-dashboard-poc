const GITHUB_URL_RE = /^https:\/\/github\.com\//
const GITHUB_AVATAR_URL_RE = /^https:\/\/avatars\.githubusercontent\.com\//

/**
 * Validates that a URL points to github.com over HTTPS.
 * Returns the URL if safe, null otherwise.
 */
export function safeGithubUrl(url: string | null | undefined): string | null {
  if (!url || !GITHUB_URL_RE.test(url)) return null
  return url
}

/**
 * Validates that an avatar URL points to avatars.githubusercontent.com over HTTPS.
 * Returns the URL if safe, null otherwise.
 */
export function safeAvatarUrl(url: string | null | undefined): string | null {
  if (!url || !GITHUB_AVATAR_URL_RE.test(url)) return null
  return url
}
