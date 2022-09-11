using FluentValidation;

namespace AuthenticatorApi.Common.Validators.User;

public class UserValidator : AbstractValidator<Models.Domain.User>
{
    public UserValidator()
    {
        RuleFor(user => user.Username).NotEmpty();
    }
}