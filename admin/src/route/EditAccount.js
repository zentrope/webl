// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import React from 'react';
import { Form, FormControls, FormWidgets, FormWidget,
         FormLabel, FormTitle } from '../component/Form'
import { WorkArea } from '../component/WorkArea'

// /admin/account/edit

const isBlank = (s) => ((! s) || (s.trim().length === 0))

class EditAccount extends React.PureComponent {

  constructor(props) {
    super(props)

    this.oldEmail = this.props.email
    this.oldName = this.props.name

    this.state = {name: this.props.name, email: this.props.email}
    this.handleChange = this.handleChange.bind(this)
    this.save = this.save.bind(this)
    this.disabled = this.disabled.bind(this)
  }

  handleChange(event) {
    let name = event.target.name
    let value = event.target.value
    this.setState({[name]: value})
  }

  save() {
    this.props.onSave(this.state.name, this.state.email)
  }

  disabled() {
    return (
      isBlank(this.state.name) ||
      isBlank(this.state.email) ||
      ((this.state.name === this.oldName) && (this.state.email === this.oldEmail))
    )
  }

  render() {
    const { onCancel } = this.props
    const { name, email } = this.state
    return (
      <WorkArea>
        <Form>
          <FormTitle>Edit account</FormTitle>
          <FormWidgets>
            <FormWidget>
              <FormLabel>Name</FormLabel>
              <input autoFocus={true} name="name" value={name} onChange={this.handleChange}/>
            </FormWidget>
            <FormWidget>
              <FormLabel>Email</FormLabel>
              <input autoFocus={false} name="email" value={email} onChange={this.handleChange}/>
            </FormWidget>
          </FormWidgets>
          <FormControls>
            <button disabled={this.disabled()} onClick={this.save}>Save changes</button>
            <button onClick={onCancel}>Cancel</button>
          </FormControls>
        </Form>
      </WorkArea>
    )
  }
}

export { EditAccount }