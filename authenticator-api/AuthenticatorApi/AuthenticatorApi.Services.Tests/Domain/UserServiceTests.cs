using AuthenticatorApi.Common.Models.Dto.Authentication;
using AuthenticatorApi.Services.Domain;

namespace AuthenticatorApi.Services.Tests.Domain;

public class UserServiceTests : DomainServiceTestBase<UserService>
{
    public override void Init()
    {
    }
    
    # region GetUserByTenantIdAndUsername
    
    [Fact]
    public async Task GetUserByTenantIdAndUsername_WhenCorrectUsernameAndTenantIdAreProvided_ReturnsUser()
    {
        await using var context = GetDbContext();
        var tenant1 = new Tenant() { Name = "tenant-1" };
        await context.AddRangeAsync(tenant1);
        await context.SaveChangesAsync();
        
        var user = new User()
        {
            Tenant = tenant1,
            Username = "test-user"
        };
        await context.AddAsync(user);
        await context.SaveChangesAsync();

        var result = await Service.GetUserByTenantIdAndUsername(tenant1.Id, user.Username);
        result.Should().NotBeNull();
        result!.Username.Should().Be(user.Username);
        result.Id.Should().Be(user.Id);
    }

    [Fact]
    public async Task GetUserByTenantIdAndUsername_WhenUserDoesNotExist_ReturnsNull()
    {
        var result = await Service.GetUserByTenantIdAndUsername(Guid.NewGuid(), "test-user");
        result.Should().BeNull();
    }
    
    [Fact]
    public async Task GetUserByTenantIdAndUsername_WhenUserDoesExistButTheWrongTenantIdIsProvided_ReturnsNull()
    {
        await using var context = GetDbContext();
        var tenant1 = new Tenant() { Name = "tenant-1" };
        var tenant2 = new Tenant() { Name = "tenant-2" };
        await context.AddRangeAsync(tenant1, tenant2);
        await context.SaveChangesAsync();
        
        var user = new User()
        {
            Tenant = tenant1,
            Username = "test-user"
        };
        await context.AddAsync(user);
        await context.SaveChangesAsync();
        
        var result = await Service.GetUserByTenantIdAndUsername(tenant2.Id, user.Username);
        result.Should().BeNull();
    }

    # endregion

    # region CreateUser

    [Fact]
    public async Task CreateUser_WhenValidDtoIsProvided_CreatesUser()
    {
        await using var context = GetDbContext();
        var tenant1 = new Tenant() { Name = "tenant-1" };
        await context.AddRangeAsync(tenant1);
        await context.SaveChangesAsync();

        var userCountBefore = await context.Users.CountAsync();
        
        var dto = new CreateUserRequestDto()
        {
            TenantId = tenant1.Id,
            Username = "test-user"
        };

        var user = await Service.CreateUser(dto);
        
        var userCountAfter = await context.Users.CountAsync();
        
        user.Should().NotBeNull();
        user.Username.Should().Be(dto.Username);
        user.IsVerified.Should().Be(false);
        user.IsActive.Should().Be(true);

        userCountBefore.Should().Be(0);
        userCountAfter.Should().Be(1);
    }
    
    [Fact]
    public async Task CreateUser_WhenTenantIdDoesNotExist_ThrowsException()
    {
        await using var context = GetDbContext();
        var tenant1 = new Tenant() { Name = "tenant-1" };
        await context.AddRangeAsync(tenant1);
        await context.SaveChangesAsync();

        var userCountBefore = await context.Users.CountAsync();
        
        var dto = new CreateUserRequestDto()
        {
            TenantId = new Guid(),
            Username = "test-user"
        };

        await Assert.ThrowsAsync<InvalidOperationException>(() => Service.CreateUser(dto));
        
        var userCountAfter = await context.Users.CountAsync();

        userCountBefore.Should().Be(0);
        userCountAfter.Should().Be(0);
    }

    # endregion
    
}