
import { AutocompleteInput, BooleanInput, Edit, TextInput, SimpleForm, ReferenceInput, useTranslate, required  } from 'react-admin';
import { RepositoryAutocomplete } from "../components/RepositoryAutocompleteField";
import RepositoryAction from "../components/RepositoryAction";

const AccessEdit = (props:any) => {
    const { source, ...rest } = props;
    const translate = useTranslate();

    return (
        <Edit title={translate('resources.groups.edit_title')}  >
            <SimpleForm>
                <TextInput sx={{ width: "30%" }} label={translate('resources.accesses.fields.name')} source="name" validate={required()} />
                <ReferenceInput source="owner_id" reference="users">
                    <AutocompleteInput sx={{ width: "30%" }} optionText="name" optionValue="id" label={translate('resources.accesses.fields.owner_id')} validate={required()}/>
                </ReferenceInput>
                <RepositoryAutocomplete source="resource_name"  validate={required()} {...rest} />
                <RepositoryAction source="action"/>
                <TextInput
                    label={translate('resources.accesses.fields.resource_type')}
                    source="type"
                    defaultValue={"repository"}
                    disabled
                />
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>

        </Edit>
    )
};

export default AccessEdit;