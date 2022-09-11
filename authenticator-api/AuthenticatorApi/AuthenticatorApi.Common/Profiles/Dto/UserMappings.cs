using AuthenticatorApi.Common.Models.Domain;
using AuthenticatorApi.Common.Models.Dto.Authentication;
using AutoMapper;

namespace AuthenticatorApi.Common.Profiles.Dto;

public class UserMappings : Profile
{
    public UserMappings()
    {
        CreateMap<CreateUserRequestDto, User>(MemberList.Source)
            .ForMember(dest => dest.IsActive, opt => opt.MapFrom(src => true));
    }
}