namespace AuthenticatorApi.Common.Models.Domain;

public class TenantAllowedDomain : AuditableEntity
{
    public Guid TenantId { get; set; }
    public Tenant Tenant { get; set; }
    public string BaseUrl { get; set; }
}