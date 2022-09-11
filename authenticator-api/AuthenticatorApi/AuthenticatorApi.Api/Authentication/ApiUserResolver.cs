using System.Security.Claims;
using AuthenticatorApi.Common.Exceptions;
using AuthenticatorApi.Common.Interfaces;

namespace AuthenticatorApi.Api.Authentication;

public class ApiUserResolver : IApplicationUserResolver
{
    private readonly IHttpContextAccessor _contextAccessor;

    public ApiUserResolver(IHttpContextAccessor contextAccessor)
    {
        _contextAccessor = contextAccessor;
    }

    public Guid GetUserId() => TryGetUserId() ?? throw new NotAuthenticatedException();

    public Guid? TryGetUserId() =>
        Guid.TryParse(_contextAccessor.HttpContext!.User?.Claims
            ?.FirstOrDefault(x => x.Type == ClaimTypes.NameIdentifier)
            ?.Value ?? "", out var parsedId)
            ? parsedId
            : (Guid?) null;
}