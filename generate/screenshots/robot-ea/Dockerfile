# Multi-arch (amd64 + arm64) Playwright image with browsers preinstalled, so
# the Enterprise Analytics robot runs natively on both Intel and Apple Silicon.
# The npm package version must match the image tag so the client and the
# preinstalled browser revisions agree.
FROM mcr.microsoft.com/playwright:v1.55.0-noble

# Uncomment this to show playwright debug output
# ENV DEBUG=pw:api

WORKDIR /src

# Install the client at build time so runs need no network access
RUN npm install --no-save playwright@1.55.0

COPY --chown=pwuser:pwuser src /src
COPY entrypoint.sh /

CMD [ "/entrypoint.sh" ]
