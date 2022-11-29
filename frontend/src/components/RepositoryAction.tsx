import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import Select from '@mui/material/Select';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import { useInput, useTranslate } from 'react-admin';


const RepositoryAction = (props:any) => {
    const translate = useTranslate();
    const {
        field,
        fieldState: { isTouched,  error },
        formState: { isSubmitted },
        isRequired
    } = useInput(props);

    return (
        <FormControl sx={{  minWidth: 180 }}>
         <InputLabel id="repository-action-select">{translate('resources.accesses.fields.action')}</InputLabel>
         <Select
            labelId='repository-action-select'
            id='repository-action-select-id'
            label={translate('resources.accesses.fields.action')}
            variant="filled"
            {...field}
            required={isRequired}
        >
            <MenuItem value="pull">Pull</MenuItem>
            <MenuItem value="push">Push</MenuItem>
        </Select>
        <FormHelperText>{(isTouched || isSubmitted) && error ? error.message : ''}</FormHelperText>
        </FormControl>
    );
};
export default RepositoryAction;
