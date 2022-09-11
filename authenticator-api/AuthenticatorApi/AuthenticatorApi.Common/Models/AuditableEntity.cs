using System.ComponentModel.DataAnnotations;

namespace AuthenticatorApi.Common.Models;

public abstract class AuditableEntity
{
    [Key]
    public Guid Id { get; set; } = Guid.NewGuid();
    public DateTimeOffset? CreatedOn { get; set; }
    public DateTimeOffset? ModifiedOn { get; set; }
    public Guid? CreatedBy { get; set; }
    public Guid? ModifiedBy { get; set; }
}