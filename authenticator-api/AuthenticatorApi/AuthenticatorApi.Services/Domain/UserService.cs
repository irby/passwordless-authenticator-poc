using AuthenticatorApi.Common.Models.Domain;
using AuthenticatorApi.Data;
using Microsoft.EntityFrameworkCore;

namespace AuthenticatorApi.Services.Domain;

public class UserService : DomainServiceBase<UserService>, IUserService
{
    public UserService(ILoggerFactory loggerFactory, ApplicationDb db) : base(loggerFactory, db)
    {
    }
    
    public async Task<User?> GetUserByTenantIdAndUsername(Guid tenantId, string username)
    {
        var users = await Db.Users.ToListAsync();
        return await Db.Users.FirstOrDefaultAsync(p => p.TenantId == tenantId && p.Username.ToLower() == username.ToLower());
    }

    public async Task CreateUser(Guid tenantId, string username)
    {
        var cleansedUsername = username.ToLower().Trim();
        var user = new User()
        {
            Username = cleansedUsername,
            FirstName = "",
            LastName = "",
            Email = "",
            IsActive = true,
            IsVerified = false,
            TenantId = tenantId
        };
        await Db.Users.AddAsync(user);
        await Db.SaveChangesAsync();
    }
}

public interface IUserService
{
    /// <summary>
    /// Reads the username from the database for the given tenant ID.
    /// </summary>
    /// <param name="tenantId">GUID of the Tenant</param>
    /// <param name="username">Username</param>
    /// <returns>A user object if the user exists for the given tenant. Returns NULL if the user does not exist for the given tenant.</returns>
    public Task<User?> GetUserByTenantIdAndUsername(Guid tenantId, string username);
}