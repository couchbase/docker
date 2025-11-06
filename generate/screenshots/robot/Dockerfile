FROM mcr.microsoft.com/playwright:v1.56.0-jammy

# Avoid redownloading bundled browsers
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

WORKDIR /src

# Install deps
COPY --chown=pwuser:pwuser package.json /src/package.json
RUN npm install --no-audit --no-fund

# Copy app and entrypoint
COPY --chown=pwuser:pwuser src /src
COPY entrypoint.sh /

CMD [ "/entrypoint.sh" ]
