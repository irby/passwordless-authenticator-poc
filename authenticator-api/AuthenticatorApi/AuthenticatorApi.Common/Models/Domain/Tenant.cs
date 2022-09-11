namespace AuthenticatorApi.Common.Models.Domain;

public class Tenant : AuditableEntity
{
    public string Name { get; set; }
    public ICollection<TenantAllowedDomain> AllowedDomains { get; set; } = new List<TenantAllowedDomain>();
}