function FindProxyForURL(url, host) {
  if (isPlainHostName(host) || shExpMatch(host, "*.local")) {
    return "DIRECT";
  }

  let websites = ["example1.com", "example2.com"];

  for (let i = 0; i < websites.length; i++) {
    if (dnsDomainIs(host, websites[i])) {
      return "SOCKS5 localhost:1080";
    }
  }

  return "DIRECT";
}
