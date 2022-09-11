namespace AuthenticatorApi.Services.Domain;

public class AuthenticationService : ServiceBase<AuthenticationService>
{
    private IUserService _userService;

    public AuthenticationService(ILoggerFactory loggerFactory, IUserService userService) : base(loggerFactory)
    {
        _userService = userService;
    }

    public async Task<bool> LoginUserAsync(Guid tenantId, string username)
    {
        var trackingId = Guid.NewGuid();
        
        Log.LogInformation($"Starting login for user {username} at tenant ID {tenantId}. Tracking ID: {trackingId}");
        
        var user = await _userService.GetUserByTenantIdAndUsername(tenantId, username);

        if (user is null)
        {
            Log.LogInformation($"Login failed for {username}. Reason: Does not exist. Tracking ID: {trackingId}");
            // Handle user being null
            return false;
        }

        if (!user.IsActive)
        {
            Log.LogInformation($"Login failed for {username}. Reason: User is not active. Tracking ID: {trackingId}");
            // Handle user deactivation
            return false;
        }

        if (!user.IsVerified)
        {
            Log.LogInformation($"Login failed for {username}. Reason: User is not yet verified. Tracking ID: {trackingId}");
            // Handle user not yet being verified
            return false;
        }
        
        Log.LogInformation($"Login successful for {username}. Tracking ID: {trackingId}");

        return true;
    }

    public async Task RegisterUser(Guid tenantId, string username)
    {
        var user = await _userService.GetUserByTenantIdAndUsername(tenantId, username);

        if (user is { })
        {
            // TODO: Handle this better
            throw new Exception();
        }
    }
}