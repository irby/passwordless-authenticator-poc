using AuthenticatorApi.Common.Models.Domain;
using AuthenticatorApi.Common.Models.Dto.Authentication;
using AuthenticatorApi.Common.Validators;
using AuthenticatorApi.Data;
using FluentValidation;
using Microsoft.EntityFrameworkCore;

namespace AuthenticatorApi.Services.Domain;

public class UserService : DomainServiceBase<UserService>, IUserService
{
    public UserService(ILoggerFactory loggerFactory, ApplicationDb db) : base(loggerFactory, db)
    {
    }
    
    public async Task<User?> GetUserByTenantIdAndUsername(Guid tenantId, string username)
    {
        return await Db.Users.FirstOrDefaultAsync(p => p.TenantId == tenantId && p.Username.ToLower() == username.ToLower());
    }

    public async Task<User> CreateUser(CreateUserRequestDto createUserRequestDto)
    {
        var user = new User()
        {
            Username = createUserRequestDto.Username,
            FirstName = createUserRequestDto.FirstName,
            LastName = createUserRequestDto.LastName,
            Email = createUserRequestDto.Email,
            IsActive = true,
            IsVerified = false,
            TenantId = createUserRequestDto.TenantId,
            PhoneNumber = createUserRequestDto.PhoneNumber
        };
        
        await new UserValidator().ValidateAndThrowAsync(user);

        await Db.Users.AddAsync(user);
        await Db.SaveChangesAsync();
        return user;
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
    Task<User?> GetUserByTenantIdAndUsername(Guid tenantId, string username);

    /// <summary>
    /// Creates a user in the database for the given username and tenant ID.
    /// </summary>
    /// <param name="createUserRequestDto">Create user DTO</param>
    Task<User> CreateUser(CreateUserRequestDto createUserRequestDto);
}