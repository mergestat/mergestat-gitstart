import { Badge, Button, Input, Label, Panel } from '@mergestat/blocks'
import { CheckIcon } from '@mergestat/icons'
import { ChangeEvent, useState } from 'react'
import { useGitSourceDetailContext } from 'src/state/contexts/git-source-detail.context'
import { showSuccessAlert } from 'src/utils/alerts'
import { LINKS_TO } from 'src/utils/constants'
import useCredentials from 'src/views/hooks/useCredentials'
import SettedAuth from './setted-auth'

const BitbucketAuth: React.FC = () => {
  const [{ gsDetail: { id, auth } }] = useGitSourceDetailContext()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  const { addCredential } = useCredentials(() => {
    showSuccessAlert('Bitbucket access token saved')
    setUsername('')
    setPassword('')
  })

  return (
    <Panel className='rounded-md shadow-sm m-auto'>
      <Panel.Header>
        <div className='w-full flex justify-between'>
          <h4 className='t-h4 mb-0'>Authentication</h4>
          {auth?.credentials &&
            <Badge
              label='Authentication is set'
              endIcon={<CheckIcon className="t-icon" />}
              variant="success"
            />}
        </div>
      </Panel.Header>

      <Panel.Body className='p-6'>
        {auth && <SettedAuth {...auth} />}

        {!auth && <>
          <p className='mb-6 t-text-muted'>
            In order to access the Bitbucket API and any private Bitbucket repos,
            we need to authenticate with an{' '}
            <a
              target='_blank'
              href={LINKS_TO.createPASS}
              className='t-link font-semibold'
              rel='noopener noreferrer'
            >
              app password
            </a>{' '}.
            Other authentication methods may become available in the future.
          </p>

          <form className='mb-4'>
            <Label>Bitbucket username</Label>
            <Input
              className='max-w-2xl'
              type='text'
              value={username}
              placeholder="Enter your Bitbucket username"
              autoComplete='off'
              onChange={(e: ChangeEvent<HTMLInputElement>) => {
                setUsername(e.target.value)
              }}
            />

            <Label className='mt-6'>Bitbucket password</Label>
            <Input
              className='max-w-2xl'
              type='password'
              value={password}
              placeholder="Enter your Bitbucket app password"
              autoComplete='off'
              onChange={(e: ChangeEvent<HTMLInputElement>) => {
                setPassword(e.target.value)
              }}
            />
          </form>

          <Button
            label='Save'
            className='mt-2'
            disabled={username === '' || password === ''}
            onClick={() => addCredential({
              variables: {
                provider: id,
                token: password,
                type: 'BITBUCKET_APP_PASSWORD',
                username
              }
            })}
          />
        </>}
      </Panel.Body>
    </Panel>
  )
}

export default BitbucketAuth
