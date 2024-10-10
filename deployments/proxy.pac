// https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file

function shuffleArray(array) {
  const shuffledArray = array.slice();

  // Fisher-Yates
  for (let i = shuffledArray.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));

    [shuffledArray[i], shuffledArray[j]] = [shuffledArray[j], shuffledArray[i]];
  }

  return shuffledArray;
}

function isInDomainList(host, domainList) {
  for (var i = 0; i < domainList.length; i++) {
    if (dnsDomainIs(host, domainList[i])) {
      return true;
    }
  }
  return false;
}

function FindProxyForURL(url, host) {
  if (isPlainHostName(host) || shExpMatch(host, "*.local")) {
    return "DIRECT";
  }

  if (shExpMatch(host, "*.a.example.com")) {
    return "PROXY a.proxy.domain:8080";
  } else if (shExpMatch(host, "*.b.example.com")) {
    return "PROXY b1.proxy.domain:8080; PROXY b2.proxy.domain:8080";
  } else if (shExpMatch(host, "*.c.example.com")) {
    return "SOCKS5 127.0.0.1:1080; SOCKS5H 127.0.0.1:1081; DIRECT";
  }

  var proxies = [
    "SOCKS5 127.0.0.1:1080",
    "SOCKS5 127.0.0.1:1081",
  ];

  return shuffleArray(proxies).join("; ");
}
