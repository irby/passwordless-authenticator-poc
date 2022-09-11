namespace AuthenticatorApi.Common.Models.Dto.Authentication;

public class CreateUserRequestDto : BaseUserRequestDto
{
    private string? _email;
    
    public string Email
    {
        get => _email;
        set => _email = value?.Trim().ToLower();
    }
    public string? FirstName { get; set; }
    public string? LastName { get; set; }
    public string? PhoneNumber { get; set; }
}