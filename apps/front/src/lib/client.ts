import { apiFetcher } from "./fetcher";

export async function clientApiFetcher<T>(
  url: string,
  params: URLSearchParams = new URLSearchParams(),
  options: RequestInit = {},
): Promise<T> {
  // The root layout sets <html lang> to the site locale.
  const locale =
    typeof document === "undefined" ? undefined : document.documentElement.lang;
  return apiFetcher<T>(url, params, options, undefined, locale);
}
