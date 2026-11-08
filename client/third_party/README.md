# WebView2Loader (Windows 7 compatible)

`WebView2Loader-x64.dll` is taken from the Microsoft.Web.WebView2 NuGet package
**1.0.1518.46**, the last WebView2 SDK release that supports Windows 7 and
Windows 8/8.1.

- Runtime support on Windows 7/8 is limited to WebView2 Runtime **109.x**.
- The DLL is redistributed under the WebView2 SDK license; see
  `WebView2Loader-LICENSE.txt` and `WebView2Loader-NOTICE.txt`.
- It is embedded into the login client and extracted at runtime to
  `%LOCALAPPDATA%\DOFLogin\WebView2Loader109-v1\WebView2Loader.dll`.
