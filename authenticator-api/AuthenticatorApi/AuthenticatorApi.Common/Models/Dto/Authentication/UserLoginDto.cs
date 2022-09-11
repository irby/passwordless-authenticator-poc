namespace AuthenticatorApi.Common.Models.Dto.Authentication;

public class UserLoginDto
{
    public string Username { get; set; }
    public Guid TenantId { get; set; }
}