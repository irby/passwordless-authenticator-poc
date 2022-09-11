namespace AuthenticatorApi.Common.Models;

public abstract class ActivatableEntity : AuditableEntity
{
    public bool IsActive { get; set; }
}