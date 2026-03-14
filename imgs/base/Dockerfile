FROM alpine:latest

# Install base dependencies and build tools
RUN apk update && apk upgrade && \
    apk add --no-cache \
    bash \
    curl \
    git \
    openssh \
    openssh-client \
    libc6-compat \
    libstdc++ \
    gcompat \
    build-base \
    linux-headers \
    jq \
    # Headless Chrome and dependencies
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    ripgrep

# Configure git
RUN git config --global user.email "overdrive@txnal.com" && \
    git config --global user.name "overdrive by transactional"

# Install opencode
RUN curl -fsSL https://opencode.ai/install | bash
ENV PATH="/root/.opencode/bin:${PATH}"

WORKDIR /overdrive
