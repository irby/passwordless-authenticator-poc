FROM mcr.microsoft.com/dotnet/sdk:6.0 AS build-env

WORKDIR /app

# Copy csproj and restore as distinct layers
COPY authenticator-api/AuthenticatorApi .
RUN dotnet restore AuthenticatorApi.sln

# Copy everything else and build
RUN dotnet publish -c Release -o artifact AuthenticatorApi.sln

# Build runtime image
FROM mcr.microsoft.com/dotnet/aspnet:6.0

WORKDIR /app
COPY --from=build-env /app/artifact .

COPY docker/docker-entrypoint.sh /usr/local/bin/

ENTRYPOINT ["sh", "/usr/local/bin/docker-entrypoint.sh"]