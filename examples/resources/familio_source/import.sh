# familio_source is imported by "<person_uuid>:<reference_uuid>" (a source is a
# sub-resource of a person, keyed by the cited entity's UUID, with no global
# address). The write-only catalog_key cannot be recovered and stays null on
# import — set it back in config for catalog_person sources.
terraform import familio_source.ivan_revision 121681c3-1234-5678-9abc-def012345678:58e68fa4-9e58-4f11-84bd-510a2dc015eb
