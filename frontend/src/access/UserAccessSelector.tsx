import {
    AutocompleteInput,
    RadioButtonGroupInput,
    ReferenceInput,
    required,
    useInput,
    useRecordContext,
    useTranslate
} from "react-admin";
import {Box, Card, CardContent, Checkbox, FormControlLabel} from '@mui/material';
import {useEffect, useState} from "react";

export const UserAccessSelector = (props: any) => {
    const translate = useTranslate();
    const record = useRecordContext();
    const [specialPermission, setSpecialPermission] = useState(false);
    const {onChange, onBlur, ...rest} = props;
    const {
        field,
        fieldState: {isTouched, error},
        formState: {isSubmitted},
    } = useInput({
        onChange,
        onBlur,
        ...props,
    });

    const choices = [
        {_id: -1000, name: 'resources.accesses.labels.label_for_all_users'},
        {_id: -999, name: 'resources.accesses.labels.label_for_registered_users'}
    ];

    useEffect(() => {
        if (record && record.owner_id && record.owner_id < 0) {
            setSpecialPermission(true);
        }
    }, [record])

    const handleParse = (value: string) => {
        return parseInt(value, 10);
    }

    return <Box sx={{width: "30%"}}>
        <Card>
            <CardContent>
                <ReferenceInput source={!specialPermission ? "owner_id" : ""} reference="users"
                                label={translate('resources.accesses.fields.owner_id')}>
                    <AutocompleteInput sx={{width: "60%"}} optionText="name" optionValue="id"
                                       disabled={specialPermission}
                                       validate={required()}/>
                </ReferenceInput>
                <FormControlLabel
                    label={translate('resources.accesses.labels.label_special_permission')}
                    control={
                        <Checkbox checked={specialPermission} onChange={(e) => setSpecialPermission(e.target.checked)}/>
                    }
                />
                <RadioButtonGroupInput
                    defaultValue={-1000}
                    parse={handleParse}
                    label=""
                    {...rest}
                    {...field}
                    disabled={!specialPermission}
                    source={specialPermission ? "owner_id" : ""}
                    choices={choices}
                    optionValue="_id"
                />

            </CardContent>
        </Card>
    </Box>
}