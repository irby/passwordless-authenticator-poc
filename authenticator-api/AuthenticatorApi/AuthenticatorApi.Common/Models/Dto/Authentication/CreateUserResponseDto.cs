namespace AuthenticatorApi.Common.Models.Dto.Authentication;

public class CreateUserResponseDto
{
    public Guid Id { get; set; }
    public bool IsVerified { get; set; }
    public DateTimeOffset CreatedOn { get; set; }
    public DateTimeOffset ModifiedOn { get; set; }
}