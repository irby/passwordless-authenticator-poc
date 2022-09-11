using AuthenticatorApi.Common.Models.Dto;
using FluentValidation;

namespace AuthenticatorApi.Common.Validators.User;

public class UserRequestDtoValidator : AbstractValidator<BaseUserRequestDto>
{
    public UserRequestDtoValidator()
    {
        RuleFor(p => p.Username).NotEmpty();
        RuleFor(p => p.TenantId).NotEmpty();
    }
}