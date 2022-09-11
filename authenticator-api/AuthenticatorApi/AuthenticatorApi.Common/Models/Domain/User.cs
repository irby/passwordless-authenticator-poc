namespace AuthenticatorApi.Common.Models.Domain;

public class User : ActivatableEntity
{
    public Guid TenantId { get; set; }
    public Tenant Tenant { get; set; }
    public string Username { get; set; }
    public string? Email { get; set; }
    public string? FirstName { get; set; }
    public string? LastName { get; set; }
    public string? PhoneNumber { get; set; }
    public Guid? ParentUserId { get; set; }
    public virtual User? ParentUser { get; set; }
    public bool IsVerified { get; set; }
    public virtual ICollection<UserPublicKey> Keys { get; set; }
}