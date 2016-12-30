FROM microsoft/windowsservercore

RUN powershell -Command \
    Invoke-WebRequest -Method Get \
   -Uri http://packages.couchbase.com/releases/4.5.1-Win10DP/couchbase-server-enterprise_4.5.1-Win10DP-windows_amd64.exe \
   -OutFile c:\couchbase.exe

COPY install.iss C:/

RUN C:\Couchbase.exe /s /f1"C:\install.iss"

RUN dir "C:\Program Files\Couchbase\Server"

COPY wait-service.ps1 C:/
CMD powershell.exe -file c:\Wait-Service.ps1 -ServiceName CouchbaseServer -AllowServiceRestart
