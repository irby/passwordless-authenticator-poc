using AuthenticatorApi.Common.Models.Domain;

namespace AuthenticatorApi.Data.Seed;

public static class SeedApplicationDb
{
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
            Id = new Guid("C4A27C5A-CEED-4E62-B607-6AB81A059786"),
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
            TenantId = new Guid("C4A27C5A-CEED-4E62-B607-6AB81A059786"),
            IsVerified = true
        });
    } 
}