using AuthenticatorApi.Common.Models.Domain;

namespace AuthenticatorApi.Data.Seed;

public static class SeedApplicationDb
{
    private static readonly Guid Tenant1Id = new Guid("C4A27C5A-CEED-4E62-B607-6AB81A059786");
    public static void Seed(this ApplicationDb context)
    {
        SeedTenants(context);
        SeedUsers(context);
        context.SaveChanges();
    }

    private static void SeedTenants(ApplicationDb context)
    {
        context.Tenants.Add(new Tenant()
        {
            Id = Tenant1Id,
            Name = "My Application",
            AllowedDomains = new List<TenantAllowedDomain>()
            {
                new TenantAllowedDomain()
                {
                    BaseUrl = "http://localhost:4200"
                }
            }
        });
    }

    private static void SeedUsers(ApplicationDb context)
    {
        context.Users.Add(new User()
        {
            Id = new Guid("D4A27C5A-CEED-4E62-B607-6AB81A059786"),
            Username = "winston",
            FirstName = "Winston",
            LastName = "Smith",
            Email = "orwell@example.com",
            IsActive = true,
            TenantId = Tenant1Id,
            IsVerified = true
        });
        context.Users.Add(new User()
        {
            Id = new Guid("D4A27C5A-CEED-4E62-B607-6AB81A059786"),
            Username = "winston1",
            FirstName = "Winston",
            LastName = "Smith",
            Email = "orwell@example.com",
            IsActive = true,
            TenantId = Tenant1Id,
            IsVerified = false
        });
        context.Users.Add(new User()
        {
            Id = new Guid("D4A27C5A-CEED-4E62-B607-6AB81A059786"),
            Username = "winston2",
            FirstName = "Winston",
            LastName = "Smith",
            Email = "orwell@example.com",
            IsActive = false,
            TenantId = Tenant1Id,
            IsVerified = true
        });
    } 
}