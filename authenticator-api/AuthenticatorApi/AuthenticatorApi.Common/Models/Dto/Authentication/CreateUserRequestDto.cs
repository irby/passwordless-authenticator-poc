namespace AuthenticatorApi.Common.Models.Dto.Authentication;

public class CreateUserRequestDto
{
    private string _username;
    private string _email;
    
    public Guid TenantId { get; set; }
    public string Username
    {
        get => _username;
        set => _username = value.Trim().ToLower();
    }
    public string Email
    {
        get => _email;
        set => _email = value.Trim().ToLower();
    }
    public string FirstName { get; set; }
    public string LastName { get; set; }
    public string PhoneNumber { get; set; }
    

}