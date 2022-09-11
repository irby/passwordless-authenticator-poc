namespace AuthenticatorApi.Common.Interfaces;

public interface IApplicationUserResolver
{
    Guid GetUserId();
    Guid? TryGetUserId();
}