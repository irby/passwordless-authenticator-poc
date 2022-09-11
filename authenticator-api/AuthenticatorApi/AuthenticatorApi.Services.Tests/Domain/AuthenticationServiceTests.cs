using AuthenticatorApi.Common.Enums;
using AuthenticatorApi.Common.Exceptions;
using AuthenticatorApi.Common.Models.Dto.Authentication;
using AuthenticatorApi.Services.Domain;
using Microsoft.Extensions.DependencyInjection;

namespace AuthenticatorApi.Services.Tests.Domain;

public class AuthenticationServiceTests : ServiceTestBase<AuthenticationService>
{
    public override void Init()
    {
        ServiceCollection.AddTransient<IUserService, UserService>();
    }

    [Fact]
    public async Task LoginUserAsync_WhenDtoIsValidAndUserExists_ProcessesRequest()
    {
        await using var context = GetDbContext();

        var tenant1 = new Tenant() { Name = "tenant1" };

        var user = new User()
        {
            Username = "test-user",
            Tenant = tenant1,
            IsActive = true,
            IsVerified = true
        };
        
        await context.AddRangeAsync(tenant1, user);
        await context.SaveChangesAsync();

        var dto = new UserLoginDto()
        {
            Username = "test-user",
            TenantId = tenant1.Id
        };
        
        await Service.LoginUserAsync(dto);
    }
    
    [Theory]
    [InlineData(" test-user ")]
    [InlineData("TEST-USER")]
    [InlineData("TeSt-UsER")]
    public async Task LoginUserAsync_WhenDtoIsValidAndVariousUsernameCasingsAreProvided_ProcessesRequest(string username)
    {
        await using var context = GetDbContext();

        var tenant1 = new Tenant() { Name = "tenant1" };

        var user = new User()
        {
            Username = "test-user",
            Tenant = tenant1,
            IsActive = true,
            IsVerified = true
        };
        
        await context.AddRangeAsync(tenant1, user);
        await context.SaveChangesAsync();

        var dto = new UserLoginDto()
        {
            Username = username,
            TenantId = tenant1.Id
        };
        
        await Service.LoginUserAsync(dto);
    }
    
    [Fact]
    public async Task LoginUserAsync_WhenDtoIsValidAndUserExistsButUserIsNotActive_ThrowsExceptionAsync()
    {
        await using var context = GetDbContext();

        var tenant1 = new Tenant() { Name = "tenant1" };

        var user = new User()
        {
            Username = "test-user",
            Tenant = tenant1,
            IsActive = false,
            IsVerified = true
        };
        
        await context.AddRangeAsync(tenant1, user);
        await context.SaveChangesAsync();

        var dto = new UserLoginDto()
        {
            Username = "test-user",
            TenantId = tenant1.Id
        };
        
        var exception = await Assert.ThrowsAsync<BadRequestException>(() => Service.LoginUserAsync(dto));
        exception.ErrorCode.Should().Be(ErrorCode.AccountDisabled);
    }
    
    [Fact]
    public async Task LoginUserAsync_WhenDtoIsValidAndUserExistsButUserIsNotVerified_ThrowsExceptionAsync()
    {
        await using var context = GetDbContext();

        var tenant1 = new Tenant() { Name = "tenant1" };

        var user = new User()
        {
            Username = "test-user",
            Tenant = tenant1,
            IsActive = true,
            IsVerified = false
        };
        
        await context.AddRangeAsync(tenant1, user);
        await context.SaveChangesAsync();

        var dto = new UserLoginDto()
        {
            Username = "test-user",
            TenantId = tenant1.Id
        };
        
        var exception = await Assert.ThrowsAsync<BadRequestException>(() => Service.LoginUserAsync(dto));
        exception.ErrorCode.Should().Be(ErrorCode.AccountNotVerified);
    }
    
    [Fact]
    public async Task LoginUserAsync_WhenDtoIsValidAndUserDoesNotExists_ThrowsExceptionAsync()
    {
        await using var context = GetDbContext();

        var tenant1 = new Tenant() { Name = "tenant1" };
        
        var user = new User()
        {
            Username = "test-user",
            Tenant = tenant1,
            IsActive = true,
            IsVerified = false
        };

        await context.AddRangeAsync(tenant1, user);
        await context.SaveChangesAsync();

        var dto = new UserLoginDto()
        {
            Username = "some-other-test-user",
            TenantId = tenant1.Id
        };
        
        await Assert.ThrowsAsync<NotFoundException>(() => Service.LoginUserAsync(dto));
    }
}