namespace AuthenticatorApi.Common.Models.Dto;

public abstract class BaseUserRequestDto
{
    private string? _username;
    
    public Guid TenantId { get; set; }
    public string Username
    {
        get => _username;
        set => _username = value?.Trim().ToLower();
    }
}