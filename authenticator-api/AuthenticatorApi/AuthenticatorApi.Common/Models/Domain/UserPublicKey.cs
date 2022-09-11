namespace AuthenticatorApi.Common.Models.Domain;

public class UserPublicKey : ActivatableEntity
{
    public Guid UserId { get; set; }
    public virtual User User { get; set; }
    public byte[] PublicKey { get; set; }
    public string Descriptor { get; set; }
}